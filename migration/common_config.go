package migration

type MysqlConfig struct {
	Addr         string `toml:"addr" json:"addr" validate:"hostname_port" long:"addr" description:"mysql server addr,format is host:port"`
	User         string `toml:"user" json:"user" long:"user" description:"mysql user"`
	Passwd       string `toml:"passwd" json:"passwd" long:"passwd" description:"mysql passwd"`
	DB           string `toml:"db" json:"db" validate:"required" long:"db" description:"mysql database name"`
	MaxOpenCount int    `toml:"max_open_count" json:"max_open_count" validate:"required" long:"max_open_count" description:"mysql connection pool max open count"`
	MaxIdleCount int    `toml:"max_idle_count" json:"max_idle_count" validate:"required" long:"max_idle_count" description:"mysql connection pool max idel count"`
	Charset      string `toml:"charset" json:"charset" long:"charset" description:"mysql charset"`
	TimeoutSec   int    `toml:"timeout_sec" json:"timeout_sec" long:"timeout_sec" description:"mysql timeout seconds"` // 超时秒数, 使用时自己拼接DSN &timeout=10s 这样
	Options      string `toml:"options" json:"options" long:"options" description:"mysql extra options, like parseTime=True&loc=Local"`
	Tracing      bool   `toml:"tracing" json:"tracing" long:"tracing" description:"enable tracing middleware"`
}

type MongoConfig struct {
	Addr        []string `toml:"addr" json:"addr" long:"addr" description:"mongo server addr,format is host:port, this option support specific multiple time" validate:"required,dive,hostname_port"`
	User        string   `toml:"user" json:"user" long:"user"`
	Passwd      string   `toml:"passwd" json:"passwd" long:"passwd"`
	AuthSource  string   `toml:"auth_source" json:"auth_source" long:"auth_source"`
	ReplicaSet  string   `toml:"replica_set" json:"replica_set" long:"replica_set"`
	DB          string   `toml:"db" json:"db" long:"db" validate:"required"`
	MinPoolSize uint64   `toml:"min_pool_size" json:"min_pool_size" long:"min_pool_size" validate:"required"`
	// MaxPoolSize The default is 100. If this is 0, it will be set to math.MaxInt64
	MaxPoolSize uint64 `toml:"max_pool_size" json:"max_pool_size" long:"max_pool_size" validate:"required"`
	// MaxConnIdleTime The default is 0, meaning a connection can remain unused indefinitely
	MaxConnIdleTime int64 `toml:"max_conn_idle_time" json:"max_conn_idle_time" long:"max_conn_idle_time"`
	// ConnectTimeout can be set through ApplyURI with the
	// "connectTimeoutMS" (e.g "connectTimeoutMS=30") option. If set to 0, no timeout will be used. The default is 30
	ConnectTimeout int64 `toml:"connect_timeout" json:"connect_timeout" long:"connect_timeout"`
	Tracing        bool  `toml:"tracing" json:"tracing" long:"tracing" description:"enable tracing middleware"`
}
