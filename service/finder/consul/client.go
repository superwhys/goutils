package consul

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/superwhys/goutils/internal/shared"
	"github.com/superwhys/goutils/lg"
	"golang.org/x/sync/singleflight"
	"gopkg.in/mgo.v2/bson"
)

type RegisteredService struct {
	ServiceID string
	CheckID   string
}

type Client struct {
	*api.Client
	sg       singleflight.Group
	services []RegisteredService
}

var HostAddress string

func IsInsideDockerContainer() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func init() {
	addrs, err := net.LookupHost("host.docker.internal")
	if err == nil && len(addrs) > 0 && IsInsideDockerContainer() {
		lg.Debugf("is inside docker container, addrs: %v", addrs)
		HostAddress = addrs[0]
		return
	}
	HostAddress = "127.0.0.1"
}

var (
	once                sync.Once
	defaultConsulClient *Client
)

func GetConsulAddress() string {
	return HostAddress + ":8500"
}

func GetConsulClient() *Client {
	once.Do(func() {
		defaultConsulClient = newConsulClient(shared.GetConsulAddress())
	})

	return defaultConsulClient
}

func newConsulClient(address string) *Client {
	config := api.DefaultConfig()
	config.Transport.Proxy = nil
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		lg.Error("Failed to connect to consul.", err)
	}
	return &Client{
		Client: client,
	}
}

func (c *Client) findInConsul(serviceName string, tag string) ([]*api.ServiceEntry, error) {
	ret, err, _ := c.sg.Do(fmt.Sprintf("%s:%s", serviceName, tag), func() (interface{}, error) {
		cs, _, err := c.Health().Service(serviceName, tag, true, nil)
		return cs, err
	})
	if err != nil {
		return nil, err
	}

	v, ok := ret.([]*api.ServiceEntry)
	if !ok {
		return nil, nil
	}
	return v, nil
}

func extractAddresses(cs []*api.ServiceEntry) []string {
	ret := make([]string, 0, len(cs))
	for _, s := range cs {
		addr := ""
		if s.Service.Address != "" {
			addr = fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port)
		} else {
			addr = fmt.Sprintf("%s:%d", s.Node.Address, s.Service.Port)
		}
		ret = append(ret, addr)
	}
	return ret
}

func (c *Client) GetAddress(service string) string {
	return c.GetAddressWithTag(service, "")
}

func (c *Client) GetAddressWithTag(service string, tag string) string {
	ip := net.ParseIP(service)
	if ip != nil {
		return ip.String()
	}

	cs := c.GetAllAddressWithTag(service, tag)
	if len(cs) == 0 {
		return ""
	}
	addr := cs[0]
	lg.Debugf("Found %s:%s -> %s in consul.", service, tag, addr)
	return addr
}

func (c *Client) GetAllAddress(service string) []string {
	return c.GetAllAddressWithTag(service, "")
}

func (c *Client) GetAllAddressWithTag(service string, tag string) []string {
	entries, err := c.findInConsul(service, tag)
	if err != nil || len(entries) == 0 {
		lg.Errorf("Failed to find %s:%s in consul.", service, tag)
		return nil
	}

	cs := extractAddresses(entries)
	rand.Shuffle(len(cs), func(i, j int) {
		cs[i], cs[j] = cs[j], cs[i]
	})

	return cs
}

var validServiceName = regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString

func (c *Client) RegisterService(serviceName string, address string) error {
	return c.RegisterServiceWithTag(serviceName, address, "")
}

func (c *Client) RegisterServiceWithTag(serviceName string, address string, tag string) error {
	if !validServiceName(serviceName) {
		return errors.New("Invalid service name")
	}

	// parse host and port from address
	ip, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = bson.NewObjectId().Hex()
	}
	hostname = strings.Replace(hostname, ".", "-", -1)
	serviceID := fmt.Sprintf("%s-%d-%s", serviceName, ip.Port, hostname)
	checkID := fmt.Sprintf("service:%s", serviceID)

	regis := &api.AgentServiceRegistration{
		ID:   serviceID,
		Name: serviceName,
		Port: ip.Port,
		Tags: []string{tag},
		Check: &api.AgentServiceCheck{
			CheckID:                        checkID,
			Name:                           serviceID,
			TCP:                            fmt.Sprintf("127.0.0.1:%d", ip.Port),
			Interval:                       (time.Second * 10).String(),
			Status:                         "passing",
			DeregisterCriticalServiceAfter: "10m",
		},
	}
	if err := c.Agent().ServiceRegister(regis); err != nil {
		return errors.Errorf("initial register service '%s' host to consul error: %s", serviceName, err.Error())
	}

	c.services = append(c.services, RegisteredService{ServiceID: serviceID, CheckID: checkID})
	return nil
}

func (c *Client) deregisterServiceAndCheck(serviceID, checkID string) (reterr error) {
	if err := c.Agent().ServiceDeregister(serviceID); err != nil {
		reterr = errors.Wrap(err, "Deregister service")
	}

	if err := c.Agent().CheckDeregister(checkID); err != nil {
		reterr = errors.Wrap(err, "Deregister check")
	}
	return
}

func (c *Client) Close() {
	for _, r := range c.services {
		if err := c.deregisterServiceAndCheck(r.ServiceID, r.CheckID); err != nil {
			lg.Error("Deregister", r.ServiceID, err)
		} else {
			lg.Info("Deregistered", r.ServiceID)
		}
	}
}
