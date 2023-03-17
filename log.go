package config

import "github.com/kk-kwok/config/version"

// log

type LogEncoding string

// the default is json, if not empty, must be one of: console|json
const (
    LogEncodingJSON    LogEncoding = "json"
    LogEncodingConsole LogEncoding = "console"
)

type LogConfig struct {
    // Level set log level, can be empty, or one of debug|info|warn|error|fatal|panic
    Level string `toml:"level" json:"level" yaml:"level"`
    // Output set output file path, can be filepath or stdout|stderr
    Output string `toml:"output" json:"output" yaml:"output"`
    // Encoding sets the logger's encoding. Valid values are "json" and "console". default: json
    Encoding LogEncoding `toml:"encoding" json:"encoding" yaml:"encoding"`
    // enable stacktrace
    DisableStacktrace bool `toml:"disable_stacktrace" json:"disable_stacktrace" yaml:"disable_stacktrace"`
}

type LoggerConfig interface {
    GetLevel() string
    GetOutput() string
    GetEncoding() string
    GetDisableStacktrace() bool
    GetInitialFields() map[string]interface{}
}

var _ LoggerConfig = &LogConfig{}

func (l *LogConfig) GetLevel() string {
    return l.Level
}

func (l *LogConfig) GetOutput() string {
    return l.Output
}

func (l *LogConfig) GetEncoding() string {
    return string(l.Encoding)
}

func (l *LogConfig) GetDisableStacktrace() bool {
    return l.DisableStacktrace
}

func (l *LogConfig) GetInitialFields() map[string]interface{} {
    return version.InitialFields()
}
