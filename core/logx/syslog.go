// +build !windows,!plan9

package logx

import (
	"log/syslog"
	"sync/atomic"
)

func setupWithSyslog(c LogConf) error {
	var err error
	once.Do(func() {
		atomic.StoreUint32(&initialized, 1)
		writeRsyslog = true
		setupLogLevel(c)

		if debugLog, err = syslog.Dial("", "", syslog.LOG_DEBUG, c.Path); err != nil {
			return
		}
		if infoLog, err = syslog.Dial("", "", syslog.LOG_INFO, c.Path); err != nil {
			return
		}
		if warnLog, err = syslog.Dial("", "", syslog.LOG_WARNING, c.Path); err != nil {
			return
		}
		if errorLog, err = syslog.Dial("", "", syslog.LOG_ERR, c.Path); err != nil {
			return
		}
		if severeLog, err = syslog.Dial("", "", syslog.LOG_CRIT, c.Path); err != nil {
			return
		}
		if slowLog, err = syslog.Dial("", "", syslog.LOG_WARNING, c.Path); err != nil {
			return
		}

		stackLog = newLessWriter(errorLog, options.logStackCooldownMills)
		statLog = debugLog // iox.NopCloser(ioutil.Discard)
	})

	return err
}
