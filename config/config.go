package config

import (
	"encoding/json"
	"fmt"
	"github.com/phayes/freeport"
	"github.com/shibukawa/configdir"
	logger "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
)

type (
	Config struct {
		NatsPort         int
		NatsHost         string
		NatsAuthRequired bool
		NatsAuthToken    string

		LogLevel string
		LogPath  string

		EveDir string
	}
)

var (
	config *Config

	configDirs = configdir.New("crypta-eve", "spyglass")

	defaultConfig = &Config{
		NatsPort: 4222,
		NatsHost: "localhost",
		LogLevel: "DEBUG",
		LogPath:  configDirs.QueryFolders(configdir.Global)[0].Path,
	}
)

func GetConfig() (cfg *Config) {
	if config == nil {
		// No config initialised, load one
		config, _ = loadConfig()
	}

	return config
}

// GetNatsPort will return the port to run the local nats server on.
// It will return -1 if no port is available
func (cfg *Config) GetNatsPort() int {
	if config.NatsPort > 0 {
		return config.NatsPort
	}
	port, err := freeport.GetFreePort()
	if err != nil {
		return -1
	}
	return port

}

func (cfg *Config) GetEveDirectory() string {
	if cfg.EveDir == "" {
		dir, err := FindEveDirectory()
		if err == nil {
			cfg.EveDir = dir
		}
	}

	return cfg.EveDir
}

func loadConfig() (*Config, error) {

	var cfg *Config

	configFolder := configDirs.QueryFolderContainsFile("settings.json")
	if configFolder != nil {
		// We have a cfg file already here read it in
		data, err := configFolder.ReadFile("settings.json")
		if err != nil {
			return defaultConfig, fmt.Errorf("failed to read settings.json: %w", err)
		}

		err = json.Unmarshal(data, cfg)
		if err != nil {
			return defaultConfig, fmt.Errorf("failed to unmarshal settings.json: %w", err)
		}
	} else {
		cfg = defaultConfig
	}

	return cfg, nil
}

func SaveConfig() error {

	data, err := json.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Stores to user folder
	folders := configDirs.QueryFolders(configdir.Global)
	err = folders[0].WriteFile("setting.json", data)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// FindEveDirectory will search the most common places or the eve log files based upon the OS in use
func FindEveDirectory() (directory string, err error) {

	directory = ""

	switch osType := runtime.GOOS; osType {

	case "darwin":
		logger.Info("OS X - Get a real OS...")
		usr, _ := user.Current()
		home := usr.HomeDir

		path := filepath.Join(home, "Documents", "EVE", "logs")
		logger.Info(path)
		if !dirExists(path) {
			logger.Warn("Directory not found")
		} else {
			files, err := ioutil.ReadDir(path)
			if err != nil {
				logger.Warn("Failed to read EVE log dir")
			} else {
				fs := make(logger.Fields)
				for index, f := range files {
					fs[fmt.Sprintf("dir%v", index)] = f.Name() + " - " + strconv.FormatBool(f.IsDir())
				}
				logger.WithFields(fs).Debug("found log dir with contents")
				directory = path
			}
		}

	case "linux":

		usr, _ := user.Current()
		home := usr.HomeDir

		path := filepath.Join(home, "Documents", "EVE", "logs")
		logger.Info(path)
		if !dirExists(path) {
			logger.Warn("Directory not found")
		} else {
			files, err := ioutil.ReadDir(path)
			if err != nil {
				logger.Fatal("Failed to read EVE log dir")
			} else {
				fs := make(logger.Fields)
				for index, f := range files {
					fs[fmt.Sprintf("dir%v", index)] = f.Name() + " - " + strconv.FormatBool(f.IsDir())
				}
				logger.WithFields(fs).Debug("found log dir with contents")
				directory = path
			}
		}

	case "windows":

		usr, _ := user.Current()
		home := usr.HomeDir

		path := filepath.Join(home, "Documents", "EVE", "logs")

		logger.Info(path)
		if !dirExists(path) {
			logger.Warn("Directory not found")
		} else {
			files, err := ioutil.ReadDir(path)
			if err != nil {
				logger.Fatal("Failed to read EVE log dir")
			} else {
				fs := make(logger.Fields)
				for index, f := range files {
					fs[fmt.Sprintf("dir%v", index)] = f.Name() + " - " + strconv.FormatBool(f.IsDir())
				}
				logger.WithFields(fs).Debug("found log dir with contents")
				directory = path
			}
		}

	default:
		logger.Fatal(fmt.Sprintf("%s", osType))

	}

	return

}

func dirExists(path string) (exists bool) {

	exists = false

	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		exists = true
	}

	return
}
