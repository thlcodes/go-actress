package log

import "fmt"

type Level int

const (
	TRACE Level = iota
	DEBUG
	INFO
	WARN
	ERROR
)

type Logger interface {
	Log(lvl Level, msg string, v ...interface{})
	Trace(msg string, v ...interface{})
	Debug(msg string, v ...interface{})
	Info(msg string, v ...interface{})
	Warn(msg string, v ...interface{})
	Error(msg string, v ...interface{})
	SubLogger(component string) Logger
}

type Field struct {
	K string
	V interface{}
}

func (f Field) String() string {
	return fmt.Sprintf("%s=%v", f.K, f.V)
}
