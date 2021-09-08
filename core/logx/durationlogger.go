package logx

import (
	"fmt"
	"io"
	"time"

	"github.com/tal-tech/go-zero/core/timex"
)

const durationCallerDepth = 3

type durationLogger logEntry

// WithDuration returns a Logger which logs the given duration.
func WithDuration(d time.Duration) Logger {
	return &durationLogger{
		Duration: timex.ReprOfDuration(d),
	}
}

func (l *durationLogger) Error(v ...interface{}) {
	if shallLog(ErrorLevel) {
		l.write(errorLog, levelError, formatWithCaller(fmt.Sprint(v...), durationCallerDepth))
	}
}

func (l *durationLogger) Errorf(format string, v ...interface{}) {
	if shallLog(ErrorLevel) {
		l.write(errorLog, levelError, formatWithCaller(fmt.Sprintf(format, v...), durationCallerDepth))
	}
}

func (l *durationLogger) Errorv(v interface{}) {
	if shallLog(ErrorLevel) {
		l.write(errorLog, levelError, v)
	}
}

func (l *durationLogger) Info(v ...interface{}) {
	if shallLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprint(v...))
	}
}

func (l *durationLogger) Infof(format string, v ...interface{}) {
	if shallLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprintf(format, v...))
	}
}

func (l *durationLogger) Infov(v interface{}) {
	if shallLog(InfoLevel) {
		l.write(infoLog, levelInfo, v)
	}
}
func (l *durationLogger) Warn(v ...interface{}) {
	if shallLog(WarnLevel) {
		l.write(warnLog, levelWarn, fmt.Sprint(v...))
	}
}

func (l *durationLogger) Warnf(format string, v ...interface{}) {
	if shallLog(WarnLevel) {
		l.write(warnLog, levelWarn, fmt.Sprintf(format, v...))
	}
}

func (l *durationLogger) Debug(v ...interface{}) {
	if shallLog(DebugLevel) {
		l.write(debugLog, levelDebug, fmt.Sprint(v...))
	}
}

func (l *durationLogger) Debugf(format string, v ...interface{}) {
	if shallLog(DebugLevel) {
		l.write(debugLog, levelDebug, fmt.Sprintf(format, v...))
	}
}

func (l *durationLogger) Slow(v ...interface{}) {
	if shallLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprint(v...))
	}
}

func (l *durationLogger) Slowf(format string, v ...interface{}) {
	if shallLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprintf(format, v...))
	}
}

func (l *durationLogger) Slowv(v interface{}) {
	if shallLog(ErrorLevel) {
		l.write(slowLog, levelSlow, v)
	}
}

func (l *durationLogger) WithDuration(duration time.Duration) Logger {
	l.Duration = timex.ReprOfDuration(duration)
	return l
}

func (l *durationLogger) write(writer io.Writer, level string, val interface{}) {
	// 1. build log entry
	if !writeRsyslog {
		l.Timestamp = getTimestamp()
		l.Level = level
	}
	l.Content = val

	// 2. output to writer
	if formatToRawTxt {
		text := fmt.Sprintf("%s %s %s: %s\n", l.Timestamp, l.Level, l.Duration, l.Content)
		outputRawTxt(writer, text)
		return
	}

	outputJson(writer, l)
}
