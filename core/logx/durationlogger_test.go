package logx

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithDurationErrorFormatToRawTxt(t *testing.T) {
	var builder strings.Builder
	log.SetOutput(&builder)
	formatToRawTxt = true
	WithDuration(time.Second).Error("foo")
	assert.True(t, strings.Contains(builder.String(), "duration"), builder.String())
	formatToRawTxt = false
}

func TestWithDurationWarn(t *testing.T) {
	var builder strings.Builder
	log.SetOutput(&builder)
	WithDuration(time.Second).Warn("foo")
	assert.True(t, strings.Contains(builder.String(), "duration"), builder.String())
}

func TestWithDurationWarnf(t *testing.T) {
	var builder strings.Builder
	log.SetOutput(&builder)
	WithDuration(time.Second).Warnf("foo")
	assert.True(t, strings.Contains(builder.String(), "duration"), builder.String())
}

func TestWithDurationDebug(t *testing.T) {
	var builder strings.Builder
	log.SetOutput(&builder)
	WithDuration(time.Second).Debug("foo")
	assert.True(t, strings.Contains(builder.String(), "duration"), builder.String())
}

func TestWithDurationDebugf(t *testing.T) {
	var builder strings.Builder
	log.SetOutput(&builder)
	WithDuration(time.Second).Debugf("foo")
	assert.True(t, strings.Contains(builder.String(), "duration"), builder.String())
}

func TestWithDurationError(t *testing.T) {
	var builder strings.Builder
	log.SetOutput(&builder)
	WithDuration(time.Second).Error("foo")
	assert.True(t, strings.Contains(builder.String(), "duration"), builder.String())
}

func TestWithDurationErrorf(t *testing.T) {
	var builder strings.Builder
	log.SetOutput(&builder)
	WithDuration(time.Second).Errorf("foo")
	assert.True(t, strings.Contains(builder.String(), "duration"), builder.String())
}

func TestWithDurationInfo(t *testing.T) {
	var builder strings.Builder
	log.SetOutput(&builder)
	WithDuration(time.Second).Info("foo")
	assert.True(t, strings.Contains(builder.String(), "duration"), builder.String())
}

func TestWithDurationInfof(t *testing.T) {
	var builder strings.Builder
	log.SetOutput(&builder)
	WithDuration(time.Second).Infof("foo")
	assert.True(t, strings.Contains(builder.String(), "duration"), builder.String())
}

func TestWithDurationSlow(t *testing.T) {
	var builder strings.Builder
	log.SetOutput(&builder)
	WithDuration(time.Second).Slow("foo")
	assert.True(t, strings.Contains(builder.String(), "duration"), builder.String())
}

func TestWithDurationSlowf(t *testing.T) {
	var builder strings.Builder
	log.SetOutput(&builder)
	WithDuration(time.Second).WithDuration(time.Hour).Slowf("foo")
	assert.True(t, strings.Contains(builder.String(), "duration"), builder.String())
}
