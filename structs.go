package config

type MysqlConfig struct {
	Dsn          string `toml:"dsn" yaml:"dsn" json:"dsn" validate:"required"` // data source name
	MaxOpenCount int    `toml:"max_open_count" json:"max_open_count" validate:"required" yaml:"max_open_count"`
	MaxIdleCount int    `toml:"max_idle_count" json:"max_idle_count" validate:"required" yaml:"max_idle_count"`
	Tracing      bool   `toml:"tracing" json:"tracing" yaml:"tracing"`
}

type GoRedisConfig struct {
	URI     string `toml:"uri" json:"uri" yaml:"uri" validate:"uri" long:"uri" description:"redis server uri"`
	Tracing bool   `toml:"tracing" json:"tracing" yaml:"tracing"`
}

type MongoConfig struct {
	DB string `toml:"db" json:"db" yaml:"db" validate:"required"`
	// URI https://www.mongodb.com/docs/manual/reference/connection-string/
	URI     string `toml:"uri" yaml:"uri" json:"uri" validate:"required"`
	Tracing bool   `toml:"tracing" json:"tracing" long:"tracing" yaml:"tracing" description:"enable tracing middleware"`
}
