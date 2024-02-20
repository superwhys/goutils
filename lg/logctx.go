package lg

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/go-logfmt/logfmt"
)

const (
	logContextKey = "logContext"
)

type LogContext struct {
	msg    []string
	keys   []string
	values []string
}

func sliceClone(strSlice []string) []string {
	if strSlice == nil {
		return strSlice
	}
	return append(strSlice[:0:0], strSlice...)
}

func cloneLogContext(c *LogContext) *LogContext {
	if c == nil {
		return nil
	}
	clone := &LogContext{
		msg:    sliceClone(c.msg),
		keys:   sliceClone(c.keys),
		values: sliceClone(c.values),
	}

	return clone
}

func parseFmtStr(format string) (msg string, isKV []bool, keys, descs []string) {
	// Format like "% d" will not be supported.
	var msgs []string
	for _, s := range strings.Split(format, " ") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		idx := strings.Index(s, "=%")
		if idx == -1 || strings.Contains(s[:idx], "=") {
			re, _ := regexp.Compile("%[^%]+")
			matches := re.FindAllStringIndex(s, -1)
			for i := 0; i < len(matches); i++ {
				isKV = append(isKV, false)
			}
			msgs = append(msgs, s)
			continue
		}
		keys = append(keys, s[:idx])
		descs = append(descs, s[idx+1:])
		isKV = append(isKV, true)
	}
	msg = strings.Join(msgs, " ")
	return
}

func With(ctx context.Context, msg string, v ...interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(msg) == 0 && len(v) == 0 {
		return ctx
	}

	lc := ParseFromContext(ctx)
	if lc == nil {
		lc = &LogContext{}
	}

	newLc := cloneLogContext(lc)

	/*
		msg= hello a=%s world %d        v = &a, 1
		hello world %d
		[true, false]
		[a]
		[%s]

		it will parse the message like a=%s out of the message and put it to the end of msg
		hello world 1 a=%s
	*/
	msgTmpl, isKv, keys, desc := parseFmtStr(msg)
	var msgV []interface{}
	var objV []interface{}

	for i, kv := range isKv {
		var val interface{}
		if i >= len(v) {
			val = "<Missing>"
		} else {
			val = v[i]
		}

		if kv {
			// a=%s
			objV = append(objV, val)
		} else {
			msgV = append(msgV, val)
		}
	}
	msg = fmt.Sprintf(msgTmpl, msgV...)
	if msg != "" {
		newLc.msg = append(newLc.msg, msg)
	}

	if len(objV) != len(desc) {
		Error("Invalid numbers of keys and values")
		return ctx
	}

	newLc.keys = append(newLc.keys, keys...)
	for i := range desc {
		newLc.values = append(newLc.values, fmt.Sprintf(desc[i], objV[i]))
	}
	return context.WithValue(ctx, logContextKey, newLc)
}

func ParseFromContext(ctx context.Context) *LogContext {
	if ctx == nil {
		return nil
	}

	val := ctx.Value(logContextKey)
	lc, ok := val.(*LogContext)
	if !ok {
		return nil
	}
	return lc
}

func (lc *LogContext) LogFmt() string {
	msg := strings.Join(lc.msg, " ")
	if len(lc.keys) != len(lc.values) {
		Error("Invalid numbers of keys and values")
		return msg
	}

	var buf bytes.Buffer

	encoder := logfmt.NewEncoder(&buf)

	for i := 0; i < len(lc.keys); i++ {
		encoder.EncodeKeyval(lc.keys[i], lc.values[i])
	}
	str := buf.String()
	if str == "" {
		return msg
	}

	return msg + " " + color.MagentaString(str)
}

type logable interface {
	Output(calldepth int, s string) error
}

func logc(ctx context.Context, l logable) {
	lc := ParseFromContext(ctx)
	if lc == nil {
		return
	}

	msg := lc.LogFmt()
	for _, line := range strings.Split(msg, "\n") {
		l.Output(3, line)
	}
}

func Infoc(ctx context.Context, msg string, v ...interface{}) {
	if len(msg) > 0 || len(v) > 0 {
		ctx = With(ctx, msg, v...)
	}
	logc(ctx, logger.infoLog)
}

func Debugc(ctx context.Context, msg string, v ...interface{}) {
	if !debug {
		return
	}

	if len(msg) > 0 || len(v) > 0 {
		ctx = With(ctx, msg, v...)
	}
	logc(ctx, logger.debugLog)
}

func Errorc(ctx context.Context, msg string, v ...interface{}) {
	if len(msg) > 0 || len(v) > 0 {
		ctx = With(ctx, msg, v...)
	}
	logc(ctx, logger.errLog)
}

func Warnc(ctx context.Context, msg string, v ...interface{}) {
	if len(msg) > 0 || len(v) > 0 {
		ctx = With(ctx, msg, v...)
	}
	logc(ctx, logger.warnLog)
}
