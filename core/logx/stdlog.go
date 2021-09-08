package logx

import "log"

type redirector struct{}

// CollectStdLog redirects system log into logx info
func CollectStdLog() {
	log.SetOutput(new(redirector))
}

func (r *redirector) Write(p []byte) (n int, err error) {
	Info(string(p))
	return len(p), nil
}
