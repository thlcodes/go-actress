package log

import (
	"io"
	stdlog "log"
)

const StdLoggerBrackets = "[]"

var stdLevelStrings = map[Level]string{
	TRACE: "TRACE",
	DEBUG: "DEBUG",
	INFO:  "INFO ",
	WARN:  "WARN ",
	ERROR: "ERROR",
}

var _ Logger = (*StdLogger)(nil)

type StdLogger struct {
	prefix string
	lvl    Level
	log    *stdlog.Logger
}

func NewStdLogger() *StdLogger {
	return &StdLogger{
		lvl: INFO,
		log: stdlog.New(stdlog.Writer(), "", stdlog.LstdFlags),
	}
}

func (sl *StdLogger) WithLevel(lvl Level) *StdLogger {
	sl.lvl = lvl
	return sl
}

func (sl *StdLogger) WithPrefix(prefix string) *StdLogger {
	return &StdLogger{
		prefix: prefix,
		log:    stdlog.New(sl.log.Writer(), "", sl.log.Flags()),
	}
}

func (sl *StdLogger) SubLogger(prefix string) Logger {
	return &StdLogger{
		prefix: sl.prefix + "|" + prefix,
		log:    stdlog.New(sl.log.Writer(), "", sl.log.Flags()),
	}
}

func (sl *StdLogger) WithOutput(w io.Writer) *StdLogger {
	return &StdLogger{
		prefix: sl.prefix,
		log:    stdlog.New(w, "", sl.log.Flags()),
	}
}

func (sl *StdLogger) WithFlags(flags int) *StdLogger {
	return &StdLogger{
		prefix: sl.prefix,
		log:    stdlog.New(sl.log.Writer(), "", flags),
	}
}

func (sl *StdLogger) Log(lvl Level, msg string, v ...interface{}) {
	if lvl < sl.lvl {
		return
	}
	level := StdLoggerBrackets[:1] + stdLevelStrings[lvl] + StdLoggerBrackets[1:]
	prefix := ""
	if sl.prefix != "" {
		prefix = " " + StdLoggerBrackets[:1] + sl.prefix + StdLoggerBrackets[1:]
	}
	sl.log.Printf(level+prefix+" "+msg, v...)
}

func (sl *StdLogger) Trace(msg string, v ...interface{}) {
	sl.Log(TRACE, msg, v...)
}

func (sl *StdLogger) Debug(msg string, v ...interface{}) {
	sl.Log(DEBUG, msg, v...)
}

func (sl *StdLogger) Info(msg string, v ...interface{}) {
	sl.Log(INFO, msg, v...)
}

func (sl *StdLogger) Warn(msg string, v ...interface{}) {
	sl.Log(WARN, msg, v...)
}

func (sl *StdLogger) Error(msg string, v ...interface{}) {
	sl.Log(ERROR, msg, v...)
}
