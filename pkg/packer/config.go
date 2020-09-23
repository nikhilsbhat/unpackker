package packer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	"github.com/nikhilsbhat/config/decode"
	"github.com/nikhilsbhat/unpackker/pkg/backend"
	"github.com/nikhilsbhat/unpackker/pkg/helper"
	"gopkg.in/yaml.v2"
)

var (
	configPathExt  = []string{".yaml", ".yml", ".json"}
	configFileName = "unpackker-config"
)

// LoadConfig will load config from file `unpackker-config` onto PackkerInput.
func (i *PackkerInput) LoadConfig() (*PackkerInput, error) {
	if err := i.getConfig(); err != nil {
		return nil, err
	}

	configContent, err := decode.ReadFile(i.ConfigPath)
	if err != nil {
		return nil, err
	}

	newConfig, err := decodeConfig(configContent, filepath.Ext(i.ConfigPath))
	if err != nil {
		return nil, err
	}
	return newConfig, nil
}

func getConfigFromEnvWithValidate() (*PackkerInput, bool, error) {
	newcfg := NewConfig()
	envcfg, bck, err := getConfigFromEnv()
	if err != nil {
		return envcfg, false, err
	}
	if (envcfg == &PackkerInput{}) && (bck == &backend.Store{}) {
		return envcfg, true, nil
	}
	newcfg = envcfg
	newcfg.Backend = bck
	return newcfg, false, nil
}

func getConfigFromEnv() (*PackkerInput, *backend.Store, error) {
	back := backend.New()
	if err := env.Parse(back); err != nil {
		return nil, nil, err
	}
	cfg := NewConfig()
	if err := env.Parse(cfg); err != nil {
		return nil, nil, err
	}
	return cfg, back, nil
}

func (i *PackkerInput) getConfig() error {
	// Environment variable for config path is set to take higher precedence.
	if len(os.Getenv("UNPACKKER_CONFIG")) != 0 {
		i.ConfigPath = os.Getenv("UNPACKKER_CONFIG")
	}

	if i.ConfigPath == "." {
		configDir, err := os.Getwd()
		if err != nil {
			return err
		}
		configPath, configexists := mapConfigFile(filepath.Join(configDir, "."+configFileName))
		if !configexists {
			return fmt.Errorf("could not find the cofig file in the current cirectory")
		}
		i.ConfigPath = configPath
		return nil
	}

	configPath, err := filepath.Abs(i.ConfigPath)
	if err != nil {
		return err
	}

	if !(validateConfigExt(configPath)) {
		return fmt.Errorf("config file format not supported, supported types are: %s", configPathExt)
	}

	if !helper.Statfile(configPath) {
		return fmt.Errorf("config file %s was not found", configPath)
	}

	i.ConfigPath = configPath
	return nil
}

func decodeConfig(rawConfig []byte, filetype string) (*PackkerInput, error) {
	config := NewConfig()

	if (filetype == ".yaml") || (filetype == ".yml") {
		if err := yaml.Unmarshal(rawConfig, &config); err != nil {
			return nil, err
		}
		return config, nil
	}
	if filetype == ".json" {
		if err := decode.JsonDecode(rawConfig, &config); err != nil {
			return nil, err
		}
		return config, nil
	}
	return nil, fmt.Errorf("failed to decode config, config file format not supported, supported types are: %s", configPathExt)
}

func mapConfigFile(filePath string) (string, bool) {
	for _, ext := range configPathExt {
		if helper.Statfile(filePath + ext) {
			return (filePath + ext), true
		}
	}
	return "", false
}

func validateConfigExt(configFile string) bool {
	for _, ext := range configPathExt {
		if filepath.Ext(configFile) == ext {
			return true
		}
	}
	return false
}
