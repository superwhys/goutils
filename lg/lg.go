package lg

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/superwhys/goutils/internal/shared"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	debug  = false
	logger *Logger
)

func init() {
	logger = New()
}

func SetDefaultLoggerOutput(stdout, stderr io.Writer) {
	logger.SetLoggerOutput(stdout, stderr)
}

func IsDebug() bool {
	return debug
}

func EnableDebug() {
	debug = true
}

func EnableLogToFile(logConf *shared.LogConfig) {
	shared.PtrLogConfig = logConf
	logger := &lumberjack.Logger{
		Filename:   logConf.FileName,
		MaxSize:    logConf.MaxSize,
		MaxBackups: logConf.MaxBackup,
		MaxAge:     logConf.MaxAge,
		Compress:   logConf.Compress,
	}

	Infof("set logger to file: %v", logConf.FileName)
	SetDefaultLoggerOutput(logger, logger)
}

func doLog(log *log.Logger, msg string) {
	for _, line := range strings.Split(msg, "\n") {
		log.Output(3, line)
	}
}

func Error(v ...interface{}) {
	if v[0] != nil {
		doLog(logger.errLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
	}
}

func PanicError(err error, msg ...interface{}) {
	var s string
	if err != nil {
		if len(msg) > 0 {
			s = err.Error() + ":" + fmt.Sprint(msg...)
		} else {
			s = err.Error()
		}
		doLog(logger.errLog, s)
		panic(err)
	}
}

func Warn(v ...interface{}) {
	if v[0] != nil {
		doLog(logger.warnLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
	}
}

func Info(v ...interface{}) {
	if v[0] != nil {
		doLog(logger.infoLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
	}
}

func Debug(v ...interface{}) {
	if debug && v[0] != nil {
		doLog(logger.debugLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
	}
}

func Fatal(v ...interface{}) {
	var msg []string
	for _, i := range v {
		msg = append(msg, fmt.Sprintf("%v", i))
	}
	doLog(logger.fatalLog, strings.Join(msg, " "))
	os.Exit(1)
}

func Jsonify(v interface{}) string {
	d, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		Error(err)
		panic(err)
	}
	return string(d)
}

func Errorf(msg string, v ...interface{}) {
	var s string
	if len(v) != 0 {
		s = strings.TrimSuffix(fmt.Sprintf(msg, v...), "\n")
	} else {
		s = msg
	}
	doLog(logger.errLog, strings.TrimSuffix(s, "\n"))

}

func Warnf(msg string, v ...interface{}) {
	var s string
	if len(v) != 0 {
		s = strings.TrimSuffix(fmt.Sprintf(msg, v...), "\n")
	} else {
		s = msg
	}
	doLog(logger.warnLog, s)
}

func Infof(msg string, v ...interface{}) {
	var s string
	if len(v) != 0 {
		s = strings.TrimSuffix(fmt.Sprintf(msg, v...), "\n")
	} else {
		s = msg
	}
	doLog(logger.infoLog, s)
}

func Debugf(msg string, v ...interface{}) {
	if !debug {
		return
	}

	var s string
	if len(v) != 0 {
		s = strings.TrimSuffix(fmt.Sprintf(msg, v...), "\n")
	} else {
		s = msg
	}
	doLog(logger.debugLog, s)
}

// TimeFuncDuration returns the duration consumed by function.
// It has specified usage like:
//
//	    f := TimeFuncDuration()
//		   DoSomething()
//		   duration := f()
func TimeFuncDuration() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Since(start)
	}
}
