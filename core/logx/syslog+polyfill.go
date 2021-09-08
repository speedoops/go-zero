// +build windows plan9

package logx

import "errors"

func setupWithSyslog(c LogConf) error {
	return errors.New("error: syslog not implemented on windows")
}
