package log

import (
	"fmt"
	"io"
	"time"
)

type TestLogger struct {
	Logger io.Writer
}

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}

func (tl TestLogger) Printf(format string, v ...interface{}) {
	fmt.Fprintf(tl.Logger, nowStamp()+": "+format+"\n", v...)
}

func (tl TestLogger) Debugf(format string, v ...interface{}) {
	fmt.Fprintf(tl.Logger, "[DEBUG] "+nowStamp()+": "+format+"\n", v...)
}
