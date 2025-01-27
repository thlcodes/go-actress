package log_test

import (
	"bytes"
	stdlog "log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thlcodes/go-actress/log"
)

func TestStdLogger(t *testing.T) {
	date := time.Now().Format("2006/01/02")
	prefix := "LOGTEST"
	buf := bytes.NewBuffer(nil)
	lut := log.NewStdLogger().WithOutput(buf).WithPrefix(prefix).WithLevel(log.TRACE).WithFlags(stdlog.Ldate | stdlog.Lmsgprefix)

	// levels prefixing
	lut.Trace("trace")
	lut.Debug("dbg")
	lut.Info("info")
	lut.Warn("warn")
	lut.Error("err")

	// level filter
	lut.WithLevel(log.INFO)
	lut.Debug("should not be logged")
	lut.Info("logged")
	lut.Error("logged")

	// sublogger
	sut1 := lut.SubLogger("sub1")
	sut1.Info("sublog1")
	sut2 := sut1.SubLogger("sub2")
	sut2.Info("sublog2")

	// message building
	lut.WithLevel(log.DEBUG)
	lut.Debug("int %d str %s", 123, "whatever")
	lut.Warn("field %s", log.Field{"A", 555})

	require.Equal(t, ""+
		date+" "+"[TRACE] [LOGTEST] trace\n"+
		date+" "+"[DEBUG] [LOGTEST] dbg\n"+
		date+" "+"[INFO ] [LOGTEST] info\n"+
		date+" "+"[WARN ] [LOGTEST] warn\n"+
		date+" "+"[ERROR] [LOGTEST] err\n"+
		date+" "+"[INFO ] [LOGTEST] logged\n"+
		date+" "+"[ERROR] [LOGTEST] logged\n"+
		date+" "+"[INFO ] [LOGTEST|sub1] sublog1\n"+
		date+" "+"[INFO ] [LOGTEST|sub1|sub2] sublog2\n"+
		date+" "+"[DEBUG] [LOGTEST] int 123 str whatever\n"+
		date+" "+"[WARN ] [LOGTEST] field A=555\n"+
		"", buf.String())
}
