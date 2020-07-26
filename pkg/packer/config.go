package packer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nikhilsbhat/config/decode"
	"gopkg.in/yaml.v2"
)

var (
	configPathExt  = []string{".yaml", ".yml", ".json"}
	configFileName = "unpackker-config"
)

// LoadConfig will load config from file `unpackker-config` onto UnpackkerInput.
func (i *UnpackkerInput) LoadConfig() (*UnpackkerInput, error) {
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

func (i *UnpackkerInput) getConfig() error {
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
		fmt.Println(configPath)
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

	if !statfile(configPath) {
		return fmt.Errorf("config file %s was not found", configPath)
	}

	i.ConfigPath = configPath
	return nil
}

func decodeConfig(rawConfig []byte, filetype string) (*UnpackkerInput, error) {
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
	return nil, fmt.Errorf("Failed to decode config, config file format not supported, supported types are: %s", configPathExt)
}

func mapConfigFile(filePath string) (string, bool) {
	for _, ext := range configPathExt {
		if statfile(filePath + ext) {
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

func statfile(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
