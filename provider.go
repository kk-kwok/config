package config

type Provider interface {
	Name() string
	Config(*providerHelper) ([]byte, error)
}

type providerHelper struct {
	configFile string
	log        Logger
}
