package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

//QscLogger gobal varirables
var QscLogger *zap.Logger

// ConfigFile profile address
var ConfigFile = "libs/logger/config.yaml"

const (
	defaultLogTimeFormat = "2006-01-02 15:04:05" //	defaultLogTimeFormat
	defaultLogMaxSize    = 1024                  // defaultLogMaxSize is the default size of log files.
	defaultLogFormat     = "json"                // defaultLogFormat is the default format of log files.
)

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

//EchoTimeEncoder set time format
func EchoTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(defaultLogTimeFormat))
}

//JSONToLogFormat format to log format
func JSONToLogFormat(JSONStr string) (map[string]interface{}, error) {
	fields := make(map[string]interface{})
	if err := json.Unmarshal([]byte(JSONStr), &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

//SetLogConigFile set config file
func SetLogConigFile(ConfigFileName string) {
	ConfigFile = ConfigFileName
}

//InitFileConfig initializes file based logging options.
func InitFileConfig(cfg *LogFileConfig) (*lumberjack.Logger, error) {
	if st, err := os.Stat(cfg.Filename); err == nil {
		if st.IsDir() {
			return nil, errors.New("can't use directory as log file name")
		}
	}

	if cfg.MaxSize == 0 {
		cfg.MaxSize = defaultLogMaxSize
	}
	now := time.Now()
	cfg.Filename = fmt.Sprintf("/tmp/%04d%02d%02d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	// use lumberjack to logrotate
	return &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxDays,
		LocalTime:  true,
	}, nil
}

// InitEncoderConfig initialize Encoder Config.
func InitEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, //lower case encoding
		EncodeTime:     EchoTimeEncoder,               //time format, eg:2006-01-02 15:04:05
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
}

// InitLoggerConfig initialize config for log file
func InitLoggerConfig(cfg *LogConfig) *zap.Config {
	atom := zap.NewAtomicLevel()

	fields := make(map[string]interface{})
	if (len(cfg.Encoding) <= 0 || cfg.Encoding == defaultLogFormat) && len(cfg.Format) > 0 {
		fields, _ = JSONToLogFormat(cfg.Format)
	}

	//init file config
	if cfg.File != nil {
		hook, err := InitFileConfig(cfg.File)
		if err != nil {
			//@todo
		}
		fmt.Println(hook)
		zapcore.AddSync(hook)
	}

	return &zap.Config{
		Level:             atom,                                  // log level
		Development:       cfg.Development,                       // development mod, stack trace
		DisableCaller:     cfg.DisableCaller,                     // close logs with the calling func
		DisableStacktrace: cfg.DisableStacktrace,                 // 关闭追踪
		Encoding:          defaultLogFormat,                      // 输出格式 console 或 json
		EncoderConfig:     InitEncoderConfig(),                   // 编码器配置
		InitialFields:     fields,                                // 初始化字段，如：添加一个服务器名称
		OutputPaths:       []string{"stdout", cfg.File.Filename}, // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths:  []string{},
	}
}

//GetLoggerConfig get log configure from local configure file
func GetLoggerConfig() (LogConfig, error) {

	viper.SetConfigName("config")      // name of config file (without extension)
	viper.AddConfigPath("libs/logger") // path to look for the config file in
	viper.AddConfigPath(".")           // optionally look for config in the working directory

	var cfg LogConfig

	// Find and read the config file
	err := viper.ReadInConfig()
	if err != nil {
		return cfg, err
	}

	err = viper.Sub("log").Unmarshal(&cfg)
	if err != nil {
		return cfg, err
	}

	logFileConfig, err := GetLogFileConfig()
	if err == nil {
		cfg.File = &logFileConfig
	}

	logFormatConfig, err := GetLogFormatConfig()
	if err == nil {
		cfg.Format = logFormatConfig
	}

	return cfg, nil
}

//GetLogFileConfig get logfile config
func GetLogFileConfig() (LogFileConfig, error) {
	var lf LogFileConfig

	err := viper.Sub("log_file").Unmarshal(&lf)
	if err != nil {
		return lf, err
	}

	return lf, nil
}

//GetLogFormatConfig get format config
func GetLogFormatConfig() (string, error) {
	var logFormatStr string
	logFormatMap := viper.Get("log_format")
	if logFormatMap != nil {
		logFormat, err := json.Marshal(logFormatMap)
		if err != nil {
			return "", err
		}
		logFormatStr = string(logFormat)
	}
	return logFormatStr, nil
}

//GetLumberJackOpts for
func GetLumberJackOpts(cfg *LogFileConfig) zapcore.WriteSyncer {
	if st, err := os.Stat(cfg.Filename); err == nil {
		if st.IsDir() {
			// return nil, errors.New("can't use directory as log file name")
		}
	}

	if cfg.MaxSize == 0 {
		cfg.MaxSize = defaultLogMaxSize
	}

	now := time.Now()
	cfg.Filename = fmt.Sprintf("/tmp/%04d%02d%02d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	// use lumberjack to logrotate
	hook := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxDays,
		LocalTime:  true,
	}

	return zapcore.AddSync(hook)
}

//NewLogger initialize logger
func NewLogger() *zap.Logger {

	cfg, err := GetLoggerConfig()
	if err != nil {
		log.Fatal("Unable to load configuration, %v", err)
	}

	zapcfg := InitLoggerConfig(&cfg)
	zapOutput := GetLumberJackOpts(cfg.File)

	lg, err := zapcfg.Build(zap.ErrorOutput(zapOutput))

	return lg
}

func init() {
	QscLogger = NewLogger()
	defer QscLogger.Sync()
}
