package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/fullstorydev/grpcui/standalone"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/soheilhy/cmux"
	"github.com/superwhys/goutils/lg"
	"github.com/superwhys/goutils/service/finder"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	welcomeText = `____  _   _ ____  _____ ____    ____  _____ ______     _____ ____ _____ 
/ ___|| | | |  _ \| ____|  _ \  / ___|| ____|  _ \ \   / /_ _/ ___| ____|
\___ \| | | | |_) |  _| | |_) | \___ \|  _| | |_) \ \ / / | | |   |  _|  
 ___) | |_| |  __/| |___|  _ <   ___) | |___|  _ < \ V /  | | |___| |___ 
|____/ \___/|_|   |_____|_| \_\ |____/|_____|_| \_\ \_/  |___\____|_____|
                                                                         `
)

type mountFn func(ctx context.Context, listener net.Listener) error
type workerFn func(ctx context.Context) error
type setterFn func(listener net.Listener) error

type workerStruct struct {
	name     string
	fullName string
	fn       workerFn
}

type SuperService struct {
	serviceName string
	tag         string

	parentCtx  context.Context
	httpCORS   bool
	withGRPCUI bool

	selfConn                  *grpc.ClientConn
	grpcServer                *grpc.Server
	grpcPreprocess            []func(*grpc.Server)
	grpcOptions               []grpc.ServerOption
	grpcGwServeMuxOption      []gwRuntime.ServeMuxOption
	unaryInterceptors         []grpc.UnaryServerInterceptor
	streamInterceptors        []grpc.StreamServerInterceptor
	grpcIncomingHeaderMapping map[string]string
	grpcOutgoingHeaderMapping map[string]string

	// cmux port multiplexing
	cmux        cmux.CMux
	httpMux     *http.ServeMux
	httpHandler http.Handler

	gatewayAPIPrefix []string
	gatewayHandlers  []gatewayFunc

	workers []*workerStruct
}

type SuperServiceOption func(*SuperService)
type gatewayFunc func(ctx context.Context, mux *gwRuntime.ServeMux, conn *grpc.ClientConn) error

func (ys *SuperService) httpIncomingHeaderMatcher(headerName string) (mdName string, ok bool) {
	if len(ys.grpcIncomingHeaderMapping) == 0 {
		return "", false
	}

	key := strings.ToLower(headerName)
	mdName, exists := ys.grpcIncomingHeaderMapping[key]
	return mdName, exists
}

func (ys *SuperService) httpOutgoingHeaderMatcher(headerName string) (mdName string, ok bool) {
	if len(ys.grpcOutgoingHeaderMapping) == 0 {
		return "", false
	}

	key := strings.ToLower(headerName)
	mdName, exists := ys.grpcOutgoingHeaderMapping[key]
	return mdName, exists
}

func WithTag(tag string) SuperServiceOption {
	return func(ms *SuperService) {
		ms.tag = tag
	}
}

func WithGrpcGatewayServeMuxOption(opt gwRuntime.ServeMuxOption) SuperServiceOption {
	return func(ys *SuperService) {
		ys.grpcGwServeMuxOption = append(ys.grpcGwServeMuxOption, opt)
	}
}

func WithGrpcOptions(opt grpc.ServerOption) SuperServiceOption {
	return func(ys *SuperService) {
		ys.grpcOptions = append(ys.grpcOptions, opt)
	}
}

func WithServiceName(name string) SuperServiceOption {
	return func(ys *SuperService) {
		lg.Debug("With service name", name)

		segs := strings.SplitN(name, ":", 2)
		if len(segs) < 2 {
			ys.serviceName = name
		} else {
			ys.serviceName = segs[0]
			ys.tag = segs[1]
		}
	}
}

func WithGRPC(preprocess func(srv *grpc.Server)) SuperServiceOption {
	return func(ys *SuperService) {
		lg.Debug("Enabled GRPC")
		ys.grpcPreprocess = append(ys.grpcPreprocess, preprocess)
	}
}

func WithGRPCUI() SuperServiceOption {
	return func(ys *SuperService) {
		ys.withGRPCUI = true
	}
}

func WithHTTPCORS() SuperServiceOption {
	return func(ys *SuperService) {
		lg.Debug("Enabled HTTP CORS")
		ys.httpCORS = true
	}
}

func WithPprof() SuperServiceOption {
	return func(ys *SuperService) {
		ys.httpMux.HandleFunc("/debug/pprof/", pprof.Index)
		ys.httpMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		ys.httpMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		ys.httpMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		ys.httpMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		ys.httpMux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		ys.httpMux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
		ys.httpMux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
		ys.httpMux.Handle("/debug/pprof/block", pprof.Handler("block"))
	}
}

func WithHttpHandler(pattern string, handler http.Handler) SuperServiceOption {
	return func(ys *SuperService) {
		if !strings.HasPrefix(pattern, "/") {
			pattern = "/" + pattern
		}

		if strings.HasSuffix(pattern, "/") {
			// pattern = pattern + "/"
			lg.Infof("Registered http endpoint prefix. prefix=%s", pattern)
			ys.httpMux.Handle(pattern, http.StripPrefix(strings.TrimSuffix(pattern, "/"), handler))
			return

		}
		lg.Debug("Added http handler for", pattern)

		ys.httpMux.Handle(pattern, http.StripPrefix(pattern, handler))
	}
}

// WithRestfulGateway binds GRPC gateway handler for all registered grpc service.
func WithRestfulGateway(apiPrefix string, handler gatewayFunc) SuperServiceOption {
	return func(ys *SuperService) {
		lg.Debug("Enabled GRPC HTTP Gateway", apiPrefix)
		prefix := strings.TrimSuffix(apiPrefix, "/")
		ys.gatewayAPIPrefix = append(ys.gatewayAPIPrefix, prefix)
		ys.gatewayHandlers = append(ys.gatewayHandlers, handler)
	}
}

// WithWorker service will terminate when any of the worker return
func WithWorker(worker func(ctx context.Context) error) SuperServiceOption {
	name := guessWorkerName(worker)
	return WithNamedWorker(name, worker)
}

// WithNamedWorker service will terminate when any of the worker return
func WithNamedWorker(name string, worker func(ctx context.Context) error) SuperServiceOption {
	return func(ys *SuperService) {
		lg.Debug(fmt.Sprintf("Added worker=%s fullname=%s", name, getFuncName(worker)))
		ys.workers = append(ys.workers, &workerStruct{
			name:     name,
			fullName: getFuncName(worker),
			fn:       worker,
		})
	}
}

func (ys *SuperService) waitHTTPServer(httpLisenter net.Listener) mountFn {
	return func(ctx context.Context, listener net.Listener) error {
		if err := http.Serve(httpLisenter, ys.httpHandler); err != nil {
			return errors.Wrap(err, "http.Serve")
		}
		return nil
	}
}

func (ys *SuperService) waitGRPCServer(grpcListener net.Listener) mountFn {
	return func(ctx context.Context, listener net.Listener) error {
		if err := ys.grpcServer.Serve(grpcListener); err != nil {
			return errors.Wrap(err, "grpcServer.Serve")
		}
		return nil
	}
}

func (ys *SuperService) waitCmux(ctx context.Context, listener net.Listener) error {
	if err := ys.cmux.Serve(); err != nil {
		return errors.Wrap(err, "cmux.Serve")
	}
	return nil
}

func (ys *SuperService) waitGraceFulKill(ctx context.Context, listener net.Listener) error {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch,
		os.Interrupt,
		syscall.SIGHUP,  // Close shell, Ctrl + D, EOF
		syscall.SIGINT,  // Ctrl + C
		syscall.SIGTERM, // supervisord default signal, `kill` without `-9`
		syscall.SIGQUIT, // Ctrl + \
		syscall.SIGKILL, // Actually it can't be caught.

	)
	select {
	case sg := <-ch:
		lg.Info("Graceful stopping server")
		if ys.selfConn != nil {
			ys.selfConn.Close()
		}
		ys.grpcServer.GracefulStop()
		lg.Info("Graceful stopped server successfully")

		return errors.Errorf("Signal: %s", sg.String())
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (ys *SuperService) mountGRPCUI(ctx context.Context, listener net.Listener) error {
	nameSplit := strings.Split(ys.serviceName, ":")
	name := nameSplit[0]

	handler, err := standalone.HandlerViaReflection(ctx, ys.selfConn, name)
	if err != nil {
		lg.Error(fmt.Sprintf("Failed to start GRPCUI: %s", err))
	} else {
		ys.httpMux.Handle("/debug/", http.StripPrefix("/debug", handler))
	}
	<-ctx.Done()
	return nil
}

func (ys *SuperService) setHTTPCORS(listener net.Listener) error {
	ys.httpHandler = cors.AllowAll().Handler(ys.httpHandler)
	return nil
}

func getFuncName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func guessWorkerName(worker func(ctx context.Context) error) string {
	str := getFuncName(worker)
	if strings.Contains(str, "/") {
		segs := strings.Split(str, "/")
		str = segs[len(segs)-1]
	}
	return str
}

func NewSuperService(opts ...SuperServiceOption) *SuperService {
	ys := &SuperService{
		parentCtx:  context.Background(),
		httpMux:    http.NewServeMux(),
		httpCORS:   true,
		withGRPCUI: false,
	}
	ys.httpHandler = ys.httpMux
	ys.unaryInterceptors = append(ys.unaryInterceptors, lg.UnaryServerInterceptor)
	ys.streamInterceptors = append(ys.streamInterceptors, lg.StreamServerInterceptor)

	for _, opt := range opts {
		opt(ys)
	}

	return ys
}

func (ys *SuperService) registerIntoConsul(ctx context.Context, listener net.Listener) error {
	addr, ok := listener.Addr().(*net.TCPAddr)
	if ok {
		if err := finder.GetServiceFinder().RegisterServiceWithTag(ys.serviceName, addr.String(), ys.tag); err != nil {
			lg.Error("Register Consul Name", err)
			return errors.Wrap(err, "Register consul name")
		}
		lg.Info("Registered", ys.serviceName)
		if len(ys.tag) > 0 {
			lg.Info("Registered with tags", ys.tag)
		}
	}

	<-ctx.Done()
	// Deregister
	finder.GetConsulServiceFinder().Close()
	return nil
}

func (ys *SuperService) mountGRPCRestfulGateway(ctx context.Context, listener net.Listener) error {
	fixGatewayVerb := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lg.Info(fmt.Sprintf("receive request: %v", r.URL.Path))
			h.ServeHTTP(w, r)
		})
	}

	opts := []gwRuntime.ServeMuxOption{
		gwRuntime.WithIncomingHeaderMatcher(ys.httpIncomingHeaderMatcher),
		gwRuntime.WithOutgoingHeaderMatcher(ys.httpOutgoingHeaderMatcher),
	}
	opts = append(opts, ys.grpcGwServeMuxOption...)
	gwmux := gwRuntime.NewServeMux(opts...)

	for i := 0; i < len(ys.gatewayHandlers); i++ {
		if err := ys.gatewayHandlers[i](ctx, gwmux, ys.selfConn); err != nil {
			lg.Error(fmt.Sprintf("Register %d gateway handler: %s", i, err.Error()))
			continue
		}
		ys.httpMux.Handle(ys.gatewayAPIPrefix[i]+"/", fixGatewayVerb(http.StripPrefix(ys.gatewayAPIPrefix[i], gwmux)))
	}
	<-ctx.Done()
	return nil
}

func (ys *SuperService) dialSelfConnection(listener net.Listener) error {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16 * 1024 * 1024)),
	}
	_, port, _ := net.SplitHostPort(listener.Addr().String())
	target := fmt.Sprintf("127.0.0.1:%s", port)
	conn, err := grpc.DialContext(ys.parentCtx, target, opts...)
	if err != nil {
		lg.Error("Failed to dial", err)
		return errors.Wrap(err, "Gateway dialing to grpc")
	}
	ys.selfConn = conn
	lg.Debug("Inited self connect")
	return nil
}

func (ys *SuperService) mountWorker(worker *workerStruct) mountFn {
	return func(ctx context.Context, listener net.Listener) error {
		err := worker.fn(ctx)
		lg.Error(fmt.Sprintf("Worker terminated error=%s", err))
		if err != nil {
			return err
		}
		return errors.Errorf("worker %s has terminated", worker.name)
	}
}

func waitContext(ctx context.Context, fn func() error) error {
	stop := make(chan error)
	go func() {
		stop <- fn()
	}()

	go func() {
		<-ctx.Done()
		lg.Debug("Worker force close after 5 seconds")
		time.Sleep(time.Second * 5)
		stop <- errors.Wrap(ctx.Err(), "Force close")
	}()

	return <-stop
}

func (ys *SuperService) Serve(listener net.Listener) error {
	if len(ys.unaryInterceptors) > 1 {
		ys.grpcOptions = append(ys.grpcOptions, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(ys.unaryInterceptors...)))
	} else if len(ys.unaryInterceptors) == 1 {
		ys.grpcOptions = append(ys.grpcOptions, grpc.UnaryInterceptor(ys.unaryInterceptors[0]))
	}

	if len(ys.streamInterceptors) > 1 {
		ys.grpcOptions = append(ys.grpcOptions, grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(ys.streamInterceptors...)))
	} else if len(ys.streamInterceptors) == 1 {
		ys.grpcOptions = append(ys.grpcOptions, grpc.StreamInterceptor(ys.streamInterceptors[0]))
	}
	ys.grpcServer = grpc.NewServer(ys.grpcOptions...)
	for _, fn := range ys.grpcPreprocess {
		fn(ys.grpcServer)
	}

	reflection.Register(ys.grpcServer)

	// port multiplexing for grpc and http
	ys.cmux = cmux.New(listener)
	grpcListener := ys.cmux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpListener := ys.cmux.Match(cmux.HTTP1Fast(), cmux.HTTP2())

	var setters []setterFn
	// the mounted function in intialized slice must be run first
	mounts := []mountFn{
		ys.waitHTTPServer(httpListener),
		ys.waitGRPCServer(grpcListener),
		ys.waitCmux,
		ys.waitGraceFulKill,
	}

	if ys.withGRPCUI {
		// dial self connection for grpcui, it must be run after cmux serve
		if err := ys.dialSelfConnection(listener); err != nil {
			return errors.Wrap(err, "Failed to dial self connection")
		}
	}

	if len(ys.serviceName) > 0 {
		mounts = append(mounts, ys.registerIntoConsul)
	}

	if len(ys.gatewayHandlers) > 0 {
		mounts = append(mounts, ys.mountGRPCRestfulGateway)
	}

	if ys.withGRPCUI {
		mounts = append(mounts, ys.mountGRPCUI)
	}

	if ys.httpCORS {
		setters = append(setters, ys.setHTTPCORS)
	}

	// service will terminate when any of the worker return
	for _, w := range ys.workers {
		mounts = append(mounts, ys.mountWorker(w))
	}

	for _, s := range setters {
		if err := s(listener); err != nil {
			return err
		}
	}

	grp, ctx := errgroup.WithContext(ys.parentCtx)
	for _, mount := range mounts {
		mount := mount
		grp.Go(func() error {
			err := waitContext(ctx, func() error {
				return mount(ctx, listener)
			})
			if err != nil {
				return err
			}
			return nil
		})
	}

	ys.displayWelcome(listener)
	if err := grp.Wait(); err != nil {
		lg.Error(fmt.Sprintf("error group error: %v", err))
		return err
	}
	return nil
}

func (ys *SuperService) displayWelcome(listener net.Listener) {
	fmt.Println(welcomeText)
	lg.Info("Listening", listener.Addr().String())
	_, port, _ := net.SplitHostPort(listener.Addr().String())

	if ys.withGRPCUI {
		lg.Info(fmt.Sprintf("GRPCUI address: http://127.0.0.1:%s/debug", port))
	}
}

func (ys *SuperService) ListenAndServer(port int) error {
	addr := ""
	if port > 0 {
		addr = fmt.Sprintf(":%d", port)
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return ys.Serve(lis)
}
