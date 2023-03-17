package config

import (
	"errors"
	"fmt"
	"os"
)

var ErrEmptyConfigFile = errors.New("error empty config file")

type FileProvider struct {
	DefaultConfigPath     string
	SkipIfPathEmpty       bool // skip this provider if config file path is empty
	SkipIfDefaultNotExist bool
}

func (p *FileProvider) Name() string {
	return "file"
}

var _ Provider = &FileProvider{}

func (p *FileProvider) Config(helper *providerHelper) ([]byte, error) {
	configFile := helper.configFile
	usingDefault := false
	if configFile == "" {
		configFile = p.DefaultConfigPath
		usingDefault = true
	}

	if configFile == "" {
		if p.SkipIfPathEmpty {
			return nil, fmt.Errorf("%w config_file=%v default_config_path=%v", ErrSkipProvider, helper.configFile, p.DefaultConfigPath)
		}
		return nil, ErrEmptyConfigFile
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) && usingDefault && p.SkipIfDefaultNotExist {
		return nil, fmt.Errorf("default config not exists, %w config_file=%v default_config_path=%v", ErrSkipProvider, helper.configFile, p.DefaultConfigPath)
	}

	fileContent, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read config from local file failed, file=%v err=%w", configFile, err)
	}
	if len(fileContent) == 0 {
		return nil, fmt.Errorf("read config from local file failed, file=%v err=%w", configFile, ErrEmptyConfig)
	}
	helper.log.Infow("read config from local file success", "config_file", configFile)
	return fileContent, nil
}
