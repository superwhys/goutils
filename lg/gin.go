package lg

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"

	logMsg = "| [CNNS] | %s %v %s | %13v | %15s | %s %4v %s| %15s | %-7s"
)

func statusCodeColor(code int) string {

	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

// MethodColor is the ANSI color for appropriately logging http method to a terminal.
func methodColor(method string) string {

	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()
		gin.Logger()

		clientIp := c.ClientIP()
		method := c.Request.Method
		if raw != "" {
			path = path + "?" + raw
		}
		statusCode := c.Writer.Status()
		spendTime := time.Now().Sub(start)

		var logFunc func(msg string, v ...interface{})
		switch {
		case statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices:
			logFunc = Infof
		case statusCode >= http.StatusMultipleChoices && statusCode < http.StatusBadRequest:
			logFunc = Warnf
		case statusCode >= http.StatusBadRequest && statusCode <= http.StatusNetworkAuthenticationRequired:
			logFunc = Errorf
		default:
			logFunc = Errorf
		}

		logFunc(
			logMsg,
			statusCodeColor(statusCode), statusCode, reset,
			spendTime,
			clientIp,
			methodColor(method), method, reset,
			path,
			c.Request.Proto,
		)
	}
}
