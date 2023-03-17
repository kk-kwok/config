package config

import (
    "errors"
    "fmt"
    "os"
    "reflect"

    "go.uber.org/zap"

    "github.com/kk-kwok/config/version"

    "github.com/spf13/pflag"
)

var ErrEmptyConfig = errors.New("got empty config")

// Base is the common config save your ass
type Base struct {
    Tracing       bool `toml:"tracing" yaml:"tracing" json:"tracing"`                      // opentelemetry tracing
    Profile       bool `toml:"profile" yaml:"profile" json:"profile"`                      // go profiling
    Metric        bool `toml:"metric" yaml:"metric" json:"metric"`                         // prometheus metrics
    MetricGo      bool `toml:"metric_go" yaml:"metric_go" json:"metric_go"`                // prometheus go metrics
    MetricProcess bool `toml:"metric_process" yaml:"metric_process" json:"metric_process"` // prometheus process metrics

    OtlpGrpcEndpoint string `toml:"otlp_grpc_endpoint" yaml:"otlp_grpc_endpoint" json:"otlp_grpc_endpoint"`

    Log LogConfig `toml:"log" yaml:"log" json:"log"`
}

type BaseConfigEmbedded interface {
    InitOtlpGrpcEndpointFromEnv()
}

func (b *Base) InitOtlpGrpcEndpointFromEnv() {
    if b.OtlpGrpcEndpoint != "" {
        return
    }
    if tmp := os.Getenv(EnvOtlpGrpcEndpoint); tmp != "" {
        b.OtlpGrpcEndpoint = tmp
    }
}

func (b *Base) LogConfig() *LogConfig {
    return &b.Log
}

type FlagParseResult interface {
    ConfigFile() string
    DumpConfig() bool
    ShowHelp() bool
    ShowVersion() bool
    Usage() func()
}

//
// func (b *Base) SentryConfig() *SentryConfig {
// 	return &b.Sentry
// }

const (
    // EnvNacosHost compatible with old one like:
    EnvNacosHost      = "NACOS_HOST"
    EnvNacosPort      = "NACOS_PORT"
    EnvNacosNamespace = "NACOS_NAMESPACE"
    EnvNacosGroup     = "NACOS_GROUP"
    EnvNacosDataID    = "NACOS_DATAID"

    EnvNacosLogLevel = "NACOS_LOG_LEVEL"

    EnvOtlpGrpcEndpoint = "OTLP_GRPC_ENDPOINT"

    FlagConfigFile = "config"
    FlagDumpConfig = "dump"

    DefaultLogLevel = "info"
    // DefaultLogLevelNacos oops, nacos logging debug log as info level
    DefaultLogLevelNacos = "error"
    DefaultLogEncoding   = LogEncodingJSON
    DefaultLogOutput     = "stderr"
)

type ConfigLoader struct {
    options    options
    configFile string
}

func New(opts ...Option) *ConfigLoader {
    options := options{
        usage:            fmt.Sprintf("Usage: %s [Options]", version.ServiceName),
        shortDescription: fmt.Sprintf("%s %s", version.ServiceName, version.Info()),
    }
    for _, o := range opts {
        o.apply(&options)
    }

    // allow override version.ServiceName via options
    if options.serviceName != "" {
        version.ServiceName = options.serviceName
    }
    // allow override version.Version via options
    if options.serviceVersion != "" {
        version.Version = options.serviceVersion
    }

    return &ConfigLoader{
        options: options,
    }
}

func (cl *ConfigLoader) Load(cfg interface{}) error {
    mtype := reflect.TypeOf(cfg)
    if mtype.Kind() != reflect.Ptr {
        return errors.New("only a pointer to struct or map can be unmarshalled from config content")
    }

    baseEmbeded, ok := cfg.(BaseConfigEmbedded)
    if !ok {
        return errors.New("error no embedded Base struct found. did your forget to embed the `infraconfig.Base` struct to your own config struct")
    }

    var flagResult FlagParseResult
    if cl.options.flagParse != nil {
        flagResult = cl.options.flagParse()
    } else {
        flagResult = cl.defaultFlagParser()
    }

    cl.configFile = flagResult.ConfigFile()

    if flagResult.ShowHelp() {
        flagResult.Usage()()
        os.Exit(0)
    }

    if flagResult.ShowVersion() {
        // nolint: forbidigo
        fmt.Println(version.Print(version.ServiceName))
        os.Exit(0)
    }

    if cl.options.logger == nil {
        zapCfg := zap.NewProductionConfig()
        zapCfg.InitialFields = map[string]interface{}{
            "service": version.ServiceName,
            "version": version.Version,
        }
        zapCfg.OutputPaths = []string{"stderr"}

        zapLogger, err := zapCfg.Build()
        if err != nil {
            panic(err)
        }
        cl.options.logger = zapLogger.Sugar()
    }

    // init logger for Load() used here
    // cl.options.logger.Info("init logger for config loader", commonLogFields...)

    // init unmarshaler
    if cl.options.unmarshaler == nil {
        cl.options.unmarshaler = TomlUnmarshaler
        cl.options.logger.Infow("using default TOML unmarshaler")
    } else {
        cl.options.logger.Infow("using custom unmarshaler")
    }

    isDump := os.Getenv("XXX_DUMP_DEMO_CFG") != "" || flagResult.DumpConfig()

    content, err := cl.getConfigViaProviders()

    if err != nil && !isDump {
        return err
    }

    err = cl.options.unmarshaler(content, cfg)
    if err != nil {
        return fmt.Errorf("unmarshal config failed, err=%w", err)
    }

    baseEmbeded.InitOtlpGrpcEndpointFromEnv()

    cl.options.logger.Infow("config loaded successfully", "config", cfg)

    if cl.options.beforeInspectHook != nil {
        cl.options.beforeInspectHook(cfg)
    }

    // for dump demo config to file
    if os.Getenv("XXX_DUMP_DEMO_CFG") != "" || flagResult.DumpConfig() {
        cl.options.logger.Infow("begin dump demo config")
        DumpDemoCfg(cfg)
        os.Exit(0)
    }

    // logging config in toml format
    if cl.options.dumpMarshalledConfig {
        cl.dumpReMarshalledConfigText(cfg)
    }

    // inspect config
    if cl.options.inspectConfig != nil {
        if err := cl.options.inspectConfig(cfg); err != nil {
            return fmt.Errorf("inspect config failed with error: %w", err)
        }
    }
    return nil
}

func (cl *ConfigLoader) dumpReMarshalledConfigText(cfg interface{}) {
    text, err := TomlMarshalIndent(cfg)
    if err == nil {
        fmt.Fprintf(os.Stderr, "--------- begin dump toml encoded config --------- :\n%s\n", text)
    } else {
        cl.options.logger.Errorw("tomlv2.Encode failed", "config", cfg)
    }
}

func (cl *ConfigLoader) getConfigViaProviders() ([]byte, error) {
    var err error
    var content []byte

    helpr := &providerHelper{
        configFile: cl.configFile,
        log:        cl.options.logger,
    }
    for _, provider := range cl.options.providers {
        content, err = provider.Config(helpr)
        if err == nil {
            break
        }
        if errors.Is(err, ErrSkipProvider) {
            cl.options.logger.Infow("config provider skipped", "provider", provider.Name(), "reason", err)
        } else {
            cl.options.logger.Errorw("try get config via provider failed", "provider", provider.Name(), "err", err)
        }
    }

    if len(cl.options.providers) == 0 {
        err = errors.New("error no config provider usable")
    }
    return content, err
}

type defaultFlagResult struct {
    configFile  string
    dumpConfig  bool
    showHelp    bool
    showVersion bool
    usage       func()
}

func (f *defaultFlagResult) ConfigFile() string {
    return f.configFile
}

func (f *defaultFlagResult) DumpConfig() bool {
    return f.dumpConfig
}

func (f *defaultFlagResult) ShowHelp() bool {
    return f.showHelp
}

func (f *defaultFlagResult) ShowVersion() bool {
    return f.showVersion
}

func (f *defaultFlagResult) Usage() func() {
    return f.usage
}

func (cl *ConfigLoader) defaultFlagParser() FlagParseResult {
    var configFile string
    var dumpConfig bool
    var showHelp, showVersion bool

    commandLine := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
    // use standalone instead of shared default pflag.CommandLine avoid "pflag redefined: config" error when unit tests
    commandLine.Usage = func() {
        fmt.Fprint(os.Stderr, cl.options.usage, "\n\n")
        fmt.Fprint(os.Stderr, cl.options.shortDescription, "\n\n")
        fmt.Fprintln(os.Stderr, commandLine.FlagUsages())
    }
    commandLine.SortFlags = false

    commandLine.StringVarP(&configFile, FlagConfigFile, "c", "", "config file path")
    commandLine.BoolVar(&dumpConfig, FlagDumpConfig, false, "dump config to toml")
    commandLine.BoolVarP(&showVersion, "version", "v", false, "display the current version of this CLI")
    commandLine.BoolVarP(&showHelp, "help", "h", false, "show help")

    if cl.options.registerFlags != nil {
        cl.options.registerFlags(commandLine)
    }

    commandLine.Parse(os.Args[1:])
    return &defaultFlagResult{configFile, dumpConfig, showHelp, showVersion, commandLine.Usage}
}
