package logx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tal-tech/go-zero/core/iox"
	"github.com/tal-tech/go-zero/core/sysx"
	"github.com/tal-tech/go-zero/core/timex"
)

const (
	// DebugLevel logs everything
	DebugLevel = iota
	// InfoLevel includes infos
	InfoLevel
	// WarnLevel includes warnings, slows
	WarnLevel
	// ErrorLevel includes errors, slows, stacks
	ErrorLevel
	// SevereLevel only log severe messages
	SevereLevel
)

const (
	debugFilename  = "debug.log"
	infoFilename   = "info.log"
	warnFilename   = "warn.log"
	errorFilename  = "error.log"
	severeFilename = "severe.log"
	slowFilename   = "slow.log"
	statFilename   = "stat.log"

	consoleMode = "console"
	volumeMode  = "volume"
	syslogMode  = "syslog"

	levelAlert  = "alert"
	levelDebug  = "debug"
	levelInfo   = "info"
	levelWarn   = "warn"
	levelError  = "error"
	levelSevere = "severe"
	levelFatal  = "fatal"
	levelSlow   = "slow"
	levelStat   = "stat"

	backupFileDelimiter = "-"
	callerInnerDepth    = 5
	flags               = 0x0
)

var (
	// ErrLogPathNotSet is an error that indicates the log path is not set.
	ErrLogPathNotSet = errors.New("log path must be set")
	// ErrLogNotInitialized is an error that log is not initialized.
	ErrLogNotInitialized = errors.New("log not initialized")
	// ErrLogServiceNameNotSet is an error that indicates that the service name is not set.
	ErrLogServiceNameNotSet = errors.New("log service name must be set")

	timeFormat   = "2006-01-02T15:04:05.000Z07"
	writeConsole bool
	writeRsyslog bool
	logLevel     uint32
	// use uint32 for atomic operations
	disableStat uint32
	debugLog    io.WriteCloser
	infoLog     io.WriteCloser
	warnLog     io.WriteCloser
	errorLog    io.WriteCloser
	severeLog   io.WriteCloser
	slowLog     io.WriteCloser
	statLog     io.WriteCloser
	stackLog    io.Writer

	once           sync.Once
	initialized    uint32
	options        logOptions
	mode           string
	formatToRawTxt bool
)

type (
	logEntry struct {
		Timestamp string      `json:"@timestamp,omitempty"`
		Level     string      `json:"level,omitempty"`
		Duration  string      `json:"duration,omitempty"`
		Content   interface{} `json:"content"`
	}

	logOptions struct {
		gzipEnabled           bool
		logStackCooldownMills int
		keepDays              int
	}

	// LogOption defines the method to customize the logging.
	LogOption func(options *logOptions)

	// A Logger represents a logger.
	Logger interface {
		Error(...interface{})
		Errorf(string, ...interface{})
		Errorv(interface{})
		Warn(...interface{})
		Warnf(string, ...interface{})
		Debug(...interface{})
		Debugf(string, ...interface{})
		Info(...interface{})
		Infof(string, ...interface{})
		Infov(interface{})
		Slow(...interface{})
		Slowf(string, ...interface{})
		Slowv(interface{})
		WithDuration(time.Duration) Logger
	}
)

// MustSetup sets up logging with given config c. It exits on error.
func MustSetup(c LogConf) {
	Must(SetUp(c))
}

// SetUp sets up the logx. If already set up, just return nil.
// we allow SetUp to be called multiple times, because for example
// we need to allow different service frameworks to initialize logx respectively.
// the same logic for SetUp
func SetUp(c LogConf) error {
	mode = c.Mode
	if len(c.TimeFormat) > 0 {
		timeFormat = c.TimeFormat
	}
	formatToRawTxt = c.FormatToRawTxt

	switch c.Mode {
	case consoleMode:
		setupWithConsole(c)
		return nil
	case syslogMode:
		return setupWithSyslog(c)
	case volumeMode:
		return setupWithVolume(c)
	default:
		return setupWithFiles(c)
	}
}

// Alert alerts v in alert level, and the message is written to error log.
func Alert(v string) {
	outputText(errorLog, levelAlert, v)
}

// Close closes the logging.
func Close() error {
	if writeConsole {
		return nil
	}

	if atomic.LoadUint32(&initialized) == 0 {
		return ErrLogNotInitialized
	}

	atomic.StoreUint32(&initialized, 0)

	if debugLog != nil {
		if err := debugLog.Close(); err != nil {
			return err
		}
	}

	if infoLog != nil {
		if err := infoLog.Close(); err != nil {
			return err
		}
	}

	if warnLog != nil {
		if err := warnLog.Close(); err != nil {
			return err
		}
	}

	if errorLog != nil {
		if err := errorLog.Close(); err != nil {
			return err
		}
	}

	if severeLog != nil {
		if err := severeLog.Close(); err != nil {
			return err
		}
	}

	if slowLog != nil {
		if err := slowLog.Close(); err != nil {
			return err
		}
	}

	if statLog != nil {
		if err := statLog.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Disable disables the logging.
func Disable() {
	once.Do(func() {
		atomic.StoreUint32(&initialized, 1)

		debugLog = iox.NopCloser(ioutil.Discard)
		infoLog = iox.NopCloser(ioutil.Discard)
		warnLog = iox.NopCloser(ioutil.Discard)
		errorLog = iox.NopCloser(ioutil.Discard)
		severeLog = iox.NopCloser(ioutil.Discard)
		slowLog = iox.NopCloser(ioutil.Discard)
		statLog = iox.NopCloser(ioutil.Discard)
		stackLog = ioutil.Discard
	})
}

// DisableStat disables the stat logs.
func DisableStat() {
	atomic.StoreUint32(&disableStat, 1)
}

// Error writes v into error log.
func Error(v ...interface{}) {
	ErrorCaller(1, v...)
}

// ErrorCaller writes v with context into error log.
func ErrorCaller(callDepth int, v ...interface{}) {
	errorTextSync(fmt.Sprint(v...), callDepth+callerInnerDepth)
}

// ErrorCallerf writes v with context in format into error log.
func ErrorCallerf(callDepth int, format string, v ...interface{}) {
	errorTextSync(fmt.Sprintf(format, v...), callDepth+callerInnerDepth)
}

// Errorf writes v with format into error log.
func Errorf(format string, v ...interface{}) {
	ErrorCallerf(1, format, v...)
}

// ErrorStack writes v along with call stack into error log.
func ErrorStack(v ...interface{}) {
	// there is newline in stack string
	stackSync(fmt.Sprint(v...))
}

// ErrorStackf writes v along with call stack in format into error log.
func ErrorStackf(format string, v ...interface{}) {
	// there is newline in stack string
	stackSync(fmt.Sprintf(format, v...))
}

// Errorv writes v into error log with json content.
// No call stack attached, because not elegant to pack the messages.
func Errorv(v interface{}) {
	errorAnySync(v)
}

func Warn(v ...interface{}) {
	warnSync(fmt.Sprint(v...), callerInnerDepth)
}

func Warnf(format string, v ...interface{}) {
	warnSync(fmt.Sprintf(format, v...), callerInnerDepth)
}

// Info writes v into access log.
func Info(v ...interface{}) {
	infoTextSync(fmt.Sprint(v...))
}

// Infof writes v with format into access log.
func Infof(format string, v ...interface{}) {
	infoTextSync(fmt.Sprintf(format, v...))
}

// Infov writes v into access log with json content.
func Infov(v interface{}) {
	infoAnySync(v)
}

func Debug(v ...interface{}) {
	debugTextSync(fmt.Sprint(v...))
}

func Debugf(format string, v ...interface{}) {
	debugTextSync(fmt.Sprintf(format, v...))
}

// Must checks if err is nil, otherwise logs the err and exits.
func Must(err error) {
	if err != nil {
		msg := formatWithCaller(err.Error(), 3)
		log.Print(msg)
		outputText(severeLog, levelFatal, msg)
		os.Exit(1)
	}
}

// SetLevel sets the logging level. It can be used to suppress some logs.
func SetLevel(level uint32) {
	atomic.StoreUint32(&logLevel, level)
}

// Severe writes v into severe log.
func Severe(v ...interface{}) {
	severeSync(fmt.Sprint(v...))
}

// Severef writes v with format into severe log.
func Severef(format string, v ...interface{}) {
	severeSync(fmt.Sprintf(format, v...))
}

// Slow writes v into slow log.
func Slow(v ...interface{}) {
	slowTextSync(fmt.Sprint(v...))
}

// Slowf writes v with format into slow log.
func Slowf(format string, v ...interface{}) {
	slowTextSync(fmt.Sprintf(format, v...))
}

// Slowv writes v into slow log with json content.
func Slowv(v interface{}) {
	slowAnySync(v)
}

// Stat writes v into stat log.
func Stat(v ...interface{}) {
	statSync(fmt.Sprint(v...))
}

// Statf writes v with format into stat log.
func Statf(format string, v ...interface{}) {
	statSync(fmt.Sprintf(format, v...))
}

// WithCooldownMillis customizes logging on writing call stack interval.
func WithCooldownMillis(millis int) LogOption {
	return func(opts *logOptions) {
		opts.logStackCooldownMills = millis
	}
}

// WithKeepDays customizes logging to keep logs with days.
func WithKeepDays(days int) LogOption {
	return func(opts *logOptions) {
		opts.keepDays = days
	}
}

// WithGzip customizes logging to automatically gzip the log files.
func WithGzip() LogOption {
	return func(opts *logOptions) {
		opts.gzipEnabled = true
	}
}

func createOutput(path string) (io.WriteCloser, error) {
	if len(path) == 0 {
		return nil, ErrLogPathNotSet
	}

	return NewLogger(path, DefaultRotateRule(path, backupFileDelimiter, options.keepDays,
		options.gzipEnabled), options.gzipEnabled)
}

func errorAnySync(v interface{}) {
	if shallLog(ErrorLevel) {
		outputAny(errorLog, levelError, v)
	}
}

func errorTextSync(msg string, callDepth int) {
	if shallLog(ErrorLevel) {
		outputError(errorLog, msg, callDepth)
	}
}

func warnSync(msg string, callDepth int) {
	if shallLog(WarnLevel) {
		outputWarn(warnLog, msg, callDepth)
	}
}

func formatWithCaller(msg string, callDepth int) string {
	var buf strings.Builder

	caller := getCaller(callDepth)
	if len(caller) > 0 {
		buf.WriteString(caller)
		buf.WriteByte(' ')
	}

	buf.WriteString(msg)

	return buf.String()
}

func getCaller(callDepth int) string {
	var buf strings.Builder

	_, file, line, ok := runtime.Caller(callDepth)
	if ok {
		pathLevel := 0
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == filepath.Separator {
				pathLevel++
				if pathLevel >= 2 {
					short = file[i+1:]
					break
				}
			}
		}

		buf.WriteString(short)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(line))
	}

	return buf.String()
}

func getTimestamp() string {
	return timex.Time().Format(timeFormat)
}

func handleOptions(opts []LogOption) {
	for _, opt := range opts {
		opt(&options)
	}
}

func infoAnySync(val interface{}) {
	if shallLog(InfoLevel) {
		outputAny(infoLog, levelInfo, val)
	}
}

func infoTextSync(msg string) {
	if shallLog(InfoLevel) {
		outputText(infoLog, levelInfo, msg)
	}
}

func debugTextSync(msg string) {
	if shallLog(DebugLevel) {
		outputText(debugLog, levelDebug, msg)
	}
}

func outputAny(writer io.Writer, level string, val interface{}) {
	info := logEntry{
		Timestamp: getTimestamp(),
		Level:     level,
		Content:   val,
	}
	outputJson(writer, info)
}

func outputText(writer io.Writer, level, msg string) {
	info := logEntry{
		// Timestamp: getTimestamp(),
		// Level:     level,
		Content: msg,
	}
	if !writeRsyslog {
		info.Timestamp = getTimestamp()
		info.Level = level
	}

	if formatToRawTxt {
		text := fmt.Sprintf("%s %s: %s\n", info.Timestamp, info.Level, info.Content)
		outputRawTxt(writer, text)
		return
	}

	outputJson(writer, info)
}

// func outputWithCaller(writer io.Writer, level, msg string, callDepth int) {
// 	content := formatWithCaller(msg, callDepth)
// 	outputText(writer, level, content)
// }

func outputRawTxt(writer io.Writer, content string) {
	if atomic.LoadUint32(&initialized) == 0 || writer == nil {
		log.Println(content)
		return
	}

	buf := append([]byte(nil), []byte(content)...)
	writer.Write(buf)
}

func outputError(writer io.Writer, msg string, callDepth int) {
	content := formatWithCaller(msg, callDepth)
	outputText(writer, levelError, content)
}

func outputWarn(writer io.Writer, msg string, callDepth int) {
	content := formatWithCaller(msg, callDepth)
	outputText(writer, levelWarn, content)
}

func outputJson(writer io.Writer, info interface{}) {
	// V1: 原始代码，对特殊字符会转义
	// if content, err := json.Marshal(info); err != nil {
	// 	log.Println(err.Error())
	// } else if atomic.LoadUint32(&initialized) == 0 || writer == nil {
	// 	log.Println(string(content))
	// } else {
	// 	writer.Write(append(content, '\n'))
	// }

	// V2: 修改后，对特殊字符不转义
	if atomic.LoadUint32(&initialized) == 0 || writer == nil {
		if content, err := json.Marshal(info); err != nil {
			log.Println(err.Error())
		} else {
			log.Println(string(content))
		}
		return
	}

	var builder strings.Builder
	enc := json.NewEncoder(&builder)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(info); err != nil {
		log.Println(err.Error())
	}
	buf := append([]byte(nil), builder.String()...)
	writer.Write(buf)
}

func setupLogLevel(c LogConf) {
	switch c.Level {
	case levelDebug:
		SetLevel(DebugLevel)
	case levelInfo:
		SetLevel(InfoLevel)
	case levelWarn:
		SetLevel(WarnLevel)
	case levelError:
		SetLevel(ErrorLevel)
	case levelSevere:
		SetLevel(SevereLevel)
	}
}

func setupWithConsole(c LogConf) {
	once.Do(func() {
		atomic.StoreUint32(&initialized, 1)
		writeConsole = true
		setupLogLevel(c)

		debugLog = newLogWriter(log.New(os.Stdout, "", flags))
		infoLog = newLogWriter(log.New(os.Stdout, "", flags))
		warnLog = newLogWriter(log.New(os.Stdout, "", flags))
		errorLog = newLogWriter(log.New(os.Stderr, "", flags))
		severeLog = newLogWriter(log.New(os.Stderr, "", flags))
		slowLog = newLogWriter(log.New(os.Stderr, "", flags))
		stackLog = newLessWriter(errorLog, options.logStackCooldownMills)
		statLog = debugLog
	})
}

func setupWithFiles(c LogConf) error {
	var opts []LogOption
	var err error

	if len(c.Path) == 0 {
		return ErrLogPathNotSet
	}

	opts = append(opts, WithCooldownMillis(c.StackCooldownMillis))
	if c.Compress {
		opts = append(opts, WithGzip())
	}
	if c.KeepDays > 0 {
		opts = append(opts, WithKeepDays(c.KeepDays))
	}

	debugFile := path.Join(c.Path, debugFilename)
	infoFile := path.Join(c.Path, infoFilename)
	warnFile := path.Join(c.Path, warnFilename)
	errorFile := path.Join(c.Path, errorFilename)
	severeFile := path.Join(c.Path, severeFilename)
	slowFile := path.Join(c.Path, slowFilename)
	statFile := path.Join(c.Path, statFilename)

	once.Do(func() {
		atomic.StoreUint32(&initialized, 1)
		handleOptions(opts)
		setupLogLevel(c)

		if debugLog, err = createOutput(debugFile); err != nil {
			return
		}

		if infoLog, err = createOutput(infoFile); err != nil {
			return
		}

		if warnLog, err = createOutput(warnFile); err != nil {
			return
		}

		if errorLog, err = createOutput(errorFile); err != nil {
			return
		}

		if severeLog, err = createOutput(severeFile); err != nil {
			return
		}

		if slowLog, err = createOutput(slowFile); err != nil {
			return
		}

		if statLog, err = createOutput(statFile); err != nil {
			return
		}

		stackLog = newLessWriter(errorLog, options.logStackCooldownMills)
	})

	return err
}

func setupWithVolume(c LogConf) error {
	if len(c.ServiceName) == 0 {
		return ErrLogServiceNameNotSet
	}

	c.Path = path.Join(c.Path, c.ServiceName, sysx.Hostname())
	return setupWithFiles(c)
}

func severeSync(msg string) {
	if shallLog(SevereLevel) {
		outputText(severeLog, levelSevere, fmt.Sprintf("%s\n%s", msg, string(debug.Stack())))
	}
}

func shallLog(level uint32) bool {
	return atomic.LoadUint32(&logLevel) <= level
}

func shallLogStat() bool {
	return atomic.LoadUint32(&disableStat) == 0
}

func slowAnySync(v interface{}) {
	if shallLog(WarnLevel) {
		outputAny(slowLog, levelSlow, v)
	}
}

func slowTextSync(msg string) {
	if shallLog(WarnLevel) {
		outputText(slowLog, levelSlow, msg)
	}
}

func stackSync(msg string) {
	if shallLog(ErrorLevel) {
		outputText(stackLog, levelError, fmt.Sprintf("%s\n%s", msg, string(debug.Stack())))
	}
}

func statSync(msg string) {
	if shallLogStat() && shallLog(InfoLevel) {
		outputText(statLog, levelStat, msg)
	}
}

type logWriter struct {
	logger *log.Logger
}

func newLogWriter(logger *log.Logger) logWriter {
	return logWriter{
		logger: logger,
	}
}

func (lw logWriter) Close() error {
	return nil
}

func (lw logWriter) Write(data []byte) (int, error) {
	lw.logger.Print(string(data))
	return len(data), nil
}
