package logger

// LogFileConfig serializes file log with config in yaml/json.
type LogFileConfig struct {
	Filename   string `mapstructure:"file_name"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxDays    int    `mapstructure:"max_days"`
	MaxBackups int    `mapstructure:"max_backups"`
}

// LogConfig serializes log with yaml/json.
type LogConfig struct {
	Level               string         `mapstructure:"level"`
	Encoding            string         `mapstructure:"encoding"`
	Format              string         `mapstructure:"format"`
	DisableTimestamp    bool           `mapstructure:"disable_timestamp" json:"disable-timestamp"`
	File                *LogFileConfig `mapstructure:"log_file"`
	Development         bool           `mapstructure:"development"`
	DisableCaller       bool           `mapstructure:"disable_caller"`
	DisableStacktrace   bool           `mapstructure:"disable_stacktrace"`
	DisableErrorVerbose bool           `mapstructure:"disable_error_verbose"`
}
