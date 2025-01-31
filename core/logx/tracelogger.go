package logx

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/tal-tech/go-zero/core/timex"
	"github.com/tal-tech/go-zero/core/trace/tracespec"
	"go.opentelemetry.io/otel/trace"
)

type traceLogger struct {
	logEntry
	Trace string `json:"trace,omitempty"`
	Span  string `json:"span,omitempty"`
	ctx   context.Context
}

func (l *traceLogger) Error(v ...interface{}) {
	if shallLog(ErrorLevel) {
		l.write(errorLog, levelError, formatWithCaller(fmt.Sprint(v...), durationCallerDepth))
	}
}

func (l *traceLogger) Errorf(format string, v ...interface{}) {
	if shallLog(ErrorLevel) {
		l.write(errorLog, levelError, formatWithCaller(fmt.Sprintf(format, v...), durationCallerDepth))
	}
}

func (l *traceLogger) Errorv(v interface{}) {
	if shallLog(ErrorLevel) {
		l.write(errorLog, levelError, v)
	}
}

func (l *traceLogger) Warn(v ...interface{}) {
	if shallLog(WarnLevel) {
		l.write(warnLog, levelWarn, formatWithCaller(fmt.Sprint(v...), durationCallerDepth))
	}
}

func (l *traceLogger) Warnf(format string, v ...interface{}) {
	if shallLog(WarnLevel) {
		l.write(warnLog, levelWarn, formatWithCaller(fmt.Sprintf(format, v...), durationCallerDepth))
	}
}

func (l *traceLogger) Info(v ...interface{}) {
	if shallLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprint(v...))
	}
}

func (l *traceLogger) Infof(format string, v ...interface{}) {
	if shallLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprintf(format, v...))
	}
}

func (l *traceLogger) Infov(v interface{}) {
	if shallLog(InfoLevel) {
		l.write(infoLog, levelInfo, v)
	}
}

func (l *traceLogger) Debug(v ...interface{}) {
	if shallLog(DebugLevel) {
		l.write(debugLog, levelDebug, fmt.Sprint(v...))
	}
}

func (l *traceLogger) Debugf(format string, v ...interface{}) {
	if shallLog(DebugLevel) {
		l.write(debugLog, levelDebug, fmt.Sprintf(format, v...))
	}
}

func (l *traceLogger) Slow(v ...interface{}) {
	if shallLog(WarnLevel) {
		l.write(slowLog, levelSlow, fmt.Sprint(v...))
	}
}

func (l *traceLogger) Slowf(format string, v ...interface{}) {
	if shallLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprintf(format, v...))
	}
}

func (l *traceLogger) Slowv(v interface{}) {
	if shallLog(ErrorLevel) {
		l.write(slowLog, levelSlow, v)
	}
}

func (l *traceLogger) WithDuration(duration time.Duration) Logger {
	l.Duration = timex.ReprOfDuration(duration)
	return l
}

func (l *traceLogger) write(writer io.Writer, level string, val interface{}) {
	// 1. build log entry
	if !writeRsyslog {
		l.Timestamp = getTimestamp()
		l.Level = level
	}
	l.Content = val
	l.Trace = traceIdFromContext(l.ctx)
	l.Span = spanIdFromContext(l.ctx)

	// 2. output to writer
	if formatToRawTxt {
		text := fmt.Sprintf("%s %s %s: %s %s %s\n", l.Timestamp, l.Level, l.Duration, l.Trace, l.Span, l.Content)
		outputRawTxt(writer, text)
		return
	}
	outputJson(writer, *l)
}

// WithContext sets ctx to log, for keeping tracing information.
func WithContext(ctx context.Context) Logger {
	return &traceLogger{
		ctx: ctx,
	}
}

func spanIdFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		return span.SpanContext().SpanID().String()
	}

	t, ok := ctx.Value(tracespec.TracingKey).(tracespec.Trace)
	if !ok {
		return ""
	}

	return t.SpanId()
}

func traceIdFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		return span.SpanContext().SpanID().String()
	}

	t, ok := ctx.Value(tracespec.TracingKey).(tracespec.Trace)
	if !ok {
		return ""
	}

	return t.TraceId()
}
