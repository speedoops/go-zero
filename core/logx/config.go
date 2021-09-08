package logx

// A LogConf is a logging config.
type LogConf struct {
	ServiceName         string `json:",optional"`
	Mode                string `json:",default=console,options=console|syslog|file|volume"`
	TimeFormat          string `json:",optional"`
	Path                string `json:",default=logs"`
	Level               string `json:",default=info,options=debug|warn|info|error|severe"`
	Compress            bool   `json:",optional"`
	KeepDays            int    `json:",optional"`
	StackCooldownMillis int    `json:",default=100"`
	FormatToRawTxt      bool   `json:",optional"`
}
