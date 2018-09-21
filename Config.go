package vutils

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type configUtils struct{}

//func (cu *configUtils) loadDevelopmentConfigSource(vse *vstore.VStoreEngine, opts Options, cwd string, isProduction bool) (error, *Config) {
//
// //in development we can load a config from the source as required or we can generate a basic one...
// //we can also load from disk too bootstrap the application...
//
//
//
//}
//
//func (cu *configUtils) loadProductionConfigSource(vse *vstore.VStoreEngine, opts Options, cwd string, isProduction bool) (error, *Config) {
//
//
//
//}
//
//func (cu *configUtils) getConfigFromSource(vse *vstore.VStoreEngine, opts Options, cwd string, isProduction bool) (error, *Config) {
//
// opts.ConfigSource = defaults.String(opts.ConfigSource, envy.Get("VANN_CONFIG_URI", "LOAD_DEFAULT"))
//
// defaultConfigKey := "DEV_CONFIG"
//
// if isProduction {
//   defaultConfigKey = ""
// }
//
// opts.ConfigKey = defaults.String(opts.ConfigKey, envy.Get("VANN_CONFIG_KEY", defaultConfigKey))
//
// if opts.ConfigSource == "LOAD_DEFAULT" {
//
//   //now iterate the config load order for default locations...
//
//
//
// } else {
//
//   //go get the config
//
// }
//
//}

func (cu *configUtils) GetConfigFromDefaultList(configID string, cwd string, defaultList []string, destinationStruct interface{}) error {

	for _, configSource := range defaultList {

		if err := cu.establishAndLoadConfigFromPath(cwd, configSource, destinationStruct); err == nil {

			return nil

		}

	}

	return errors.New(fmt.Sprintf("Unable to locate the required config %s at any of the supplied locations.", configID))

}

func (cu *configUtils) establishAndLoadConfigFromPath(cwd string, configSource string, destinationStruct interface{}) error {

	if strings.HasPrefix(configSource, "./") {

		//relative path config source...

		if fullPath := filepath.Join(cwd, configSource[2:]); Files.CheckPathExists(fullPath) {

			return cu.loadConfigFromFile(fullPath, destinationStruct)

		} else {

			return errors.New(fmt.Sprintf("Unable to load config from source %s as it doesnt exist in CWD or RootFolder.", configSource))

		}

	} else if strings.HasPrefix(configSource, ".") {

		//.file config source

		if fullPath := filepath.Join(cwd, configSource); Files.CheckPathExists(fullPath) {

			return cu.loadConfigFromFile(fullPath, destinationStruct)

		} else {

			return errors.New(fmt.Sprintf("Unable to load config from source %s as it doesnt exist in CWD or RootFolder.", configSource))

		}

	} else if strings.HasPrefix(configSource, "/") {

		//full path config source...

		return cu.loadConfigFromFile(configSource, destinationStruct)

	}

	return errors.New(fmt.Sprintf("Unable to load config from source %s as the source doesn't exist.", configSource))

}

func (cu *configUtils) loadConfigFromFile(path string, destinationStruct interface{}) error {

	if !Files.CheckPathExists(path) {

		return errors.New(fmt.Sprintf("Unable to load config from %s", path))

	} else {

		if contents, err := ioutil.ReadFile(path); err == nil {

			if err = json.Unmarshal(contents, destinationStruct); err == nil {

				return nil

			} else {

				return err

			}

		} else {

			return err

		}

	}

}

func (cu *configUtils) writeConfigToFile(path string, conf interface{}) error {

	if encConf, err := json.Marshal(conf); err != nil {

		return err

	} else if err := ioutil.WriteFile(path, encConf, 0640); err != nil {

		return err

	}

	return nil

}

func (cu *configUtils) SaveConfigToFile(cwd string, path string, conf interface{}) (error, string) {

	if strings.HasPrefix(path, "./") {

		confPath := filepath.Join(cwd, path)

		if err := cu.writeConfigToFile(confPath, conf); err == nil {

			return nil, confPath

		} else {

			return err, ""

		}

	} else if strings.HasPrefix(path, ".") {

		confPath := filepath.Join(cwd, path)

		if err := cu.writeConfigToFile(confPath, conf); err == nil {

			return nil, confPath

		} else {

			return err, ""

		}

	} else if strings.HasPrefix(path, "/") {

		if err := cu.writeConfigToFile(path, conf); err == nil {

			return nil, path

		} else {

			return err, ""

		}

	}

	return errors.New(fmt.Sprintf("Unable to save initial VStoreCore config to disk.")), ""

}

func (cu *configUtils) TrySaveConfig(cwd string, defaultList []string, conf interface{}) (error, string) {

	for _, loc := range defaultList {

		if err, path := cu.SaveConfigToFile(cwd, loc, conf); err == nil {

			return nil, path

		}

	}

	return errors.New(fmt.Sprintf("Unable to save initial VStoreCore Initial config to disk.")), ""

}

var Config = &configUtils{}
