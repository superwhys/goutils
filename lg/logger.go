package lg

import (
	"io"
	"log"
	"os"

	"github.com/fatih/color"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	infoLog  *log.Logger
	debugLog *log.Logger
	warnLog  *log.Logger
	errLog   *log.Logger
	fatalLog *log.Logger
}

type Option func(*Logger)

func WithFileOption(filename string, maxSize, maxBackup, maxAge int, logCompress bool) Option {
	return func(l *Logger) {
		logger := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    maxSize,
			MaxBackups: maxBackup,
			MaxAge:     maxAge,
			Compress:   logCompress,
		}
		l.SetLoggerOutput(logger, logger)
	}
}

func New(options ...Option) *Logger {
	var stdout io.Writer = os.Stdout
	var stderr io.Writer = os.Stderr

	l := &Logger{
		infoLog:  log.New(stdout, color.GreenString("[INFO]"), log.LstdFlags|log.LUTC),
		debugLog: log.New(stdout, color.CyanString("[DEBUG]"), log.LstdFlags|log.Lshortfile|log.LUTC),
		errLog:   log.New(stderr, color.RedString("[ERROR]"), log.LstdFlags|log.Lshortfile|log.LUTC),
		warnLog:  log.New(stdout, color.YellowString("[WARN]"), log.LstdFlags|log.LUTC),
		fatalLog: log.New(stderr, color.RedString("[FATAL]"), log.LstdFlags|log.Llongfile|log.LUTC),
	}

	for _, opt := range options {
		opt(l)
	}

	return l
}

func (l *Logger) SetLoggerOutput(stdout, stderr io.Writer) {
	l.infoLog.SetOutput(stdout)
	l.debugLog.SetOutput(stdout)
	l.errLog.SetOutput(stderr)
	l.warnLog.SetOutput(stdout)
	l.fatalLog.SetOutput(stderr)
}
