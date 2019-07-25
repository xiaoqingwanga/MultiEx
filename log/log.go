package log

import (
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Initlog init log for use, should init only once.
func Initlog(logLevel string, logTo string) {
	// Lower input
	logTo = strings.ToLower(logTo)
	logLevel = strings.ToLower(logLevel)

	var writer io.Writer
	switch logTo {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		var err error
		writer, err = os.Create(logTo)
		if err != nil {
			panic(err)
		}
	}
	log.SetOutput(writer)

	switch logLevel {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn", "warnning":
		log.SetLevel(log.WarnLevel)
	case "err", "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	default:
		panic("log level:" + logLevel + ",not supported")
	}
}

// Debug log, use prefix logger instead.
func Debug(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Debugf(format, args...)
	} else {
		log.Debug(format)
	}
}

// Info log, use prefix logger instead.
func Info(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Infof(format, args...)
	} else {
		log.Info(format)
	}
}

// Warn log, use prefix logger instead.
func Warn(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Warnf(format, args...)
	} else {
		log.Warn(format)
	}
}

// Error log, use prefix logger instea.
func Error(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Errorf(format, args...)
	} else {
		log.Error(args)
	}
}

// PrefixLogger is a logger with prefix
type PrefixLogger struct {
	prefix string
}

// NewPrefixLogger set prefixs for logger
func NewPrefixLogger(pfx ...string) PrefixLogger {
	pl := PrefixLogger{}
	for _, p := range pfx {
		pl.AddPrefix(p)
	}
	return pl
}

// AddPrefix add prefix to log.
func (pf *PrefixLogger) AddPrefix(pfx string) {
	if len(pf.prefix) != 0 {
		pf.prefix += " "
	}

	pf.prefix += "[" + pfx + "]"
}

func (pf *PrefixLogger) ReplacePrefix(old string,new string) {
	pf.prefix = strings.ReplaceAll(pf.prefix,old,new)
}

// Debug log
func (pf *PrefixLogger) Debug(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Debugf(pf.prefix+" "+format, args...)
	} else {
		log.Debug(pf.prefix + " " + format)
	}
}

// Info log
func (pf *PrefixLogger) Info(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Infof(pf.prefix+" "+format, args...)
	} else {
		log.Info(pf.prefix + " " + format)
	}
}

// Warn log
func (pf *PrefixLogger) Warn(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Warnf(pf.prefix+" "+format, args...)
	} else {
		log.Warn(pf.prefix + " " + format)
	}
}

// Error log
func (pf *PrefixLogger) Error(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Errorf(pf.prefix+" "+format, args...)
	} else {
		log.Error(pf.prefix + " " + format)
	}
}
