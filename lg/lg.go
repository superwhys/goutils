package lg

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	debug                                        = false
	infoLog, debugLog, warnLog, errLog, fatalLog *log.Logger
)

func init() {
	infoLog = log.New(os.Stdout, color.GreenString("[INFO]"), log.LstdFlags|log.LUTC)
	debugLog = log.New(os.Stdout, color.CyanString("[DEBUG]"), log.LstdFlags|log.Lshortfile|log.LUTC)
	errLog = log.New(os.Stderr, color.RedString("[ERROR]"), log.LstdFlags|log.Lshortfile|log.LUTC)
	warnLog = log.New(os.Stdout, color.YellowString("[WARN]"), log.LstdFlags|log.LUTC)
	fatalLog = log.New(os.Stderr, color.RedString("[FATAL]"), log.LstdFlags|log.Llongfile|log.LUTC)
}

func IsDebug() bool {
	return debug
}

func EnableDebug() {
	debug = true
}

func doLog(log *log.Logger, msg string) {
	for _, line := range strings.Split(msg, "\n") {
		log.Output(3, line)
	}
}

func Error(v ...interface{}) {
	if v[0] != nil {
		doLog(errLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
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
		doLog(errLog, s)
		panic(err)
	}
}

func Warn(v ...interface{}) {
	if v[0] != nil {
		doLog(warnLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
	}
}

func Info(v ...interface{}) {
	if v[0] != nil {
		doLog(infoLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
	}
}

func Debug(v ...interface{}) {
	if debug && v[0] != nil {
		doLog(debugLog, strings.TrimSuffix(fmt.Sprintln(v...), "\n"))
	}
}

func Fatal(v ...interface{}) {
	var msg []string
	for _, i := range v {
		msg = append(msg, fmt.Sprintf("%v", i))
	}
	doLog(fatalLog, strings.Join(msg, " "))
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
	doLog(errLog, strings.TrimSuffix(s, "\n"))

}

func Warnf(msg string, v ...interface{}) {
	var s string
	if len(v) != 0 {
		s = strings.TrimSuffix(fmt.Sprintf(msg, v...), "\n")
	} else {
		s = msg
	}
	doLog(warnLog, s)
}

func Infof(msg string, v ...interface{}) {
	var s string
	if len(v) != 0 {
		s = strings.TrimSuffix(fmt.Sprintf(msg, v...), "\n")
	} else {
		s = msg
	}
	doLog(infoLog, s)
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
	doLog(debugLog, s)
}

// TimeFuncDuration returns the duration consumed by function.
// It has specified usage like:
//     f := TimeFuncDuration()
//	   DoSomething()
//	   duration := f()
func TimeFuncDuration() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Since(start)
	}
}
