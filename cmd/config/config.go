// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"

	"3e8.eu/go/dsl"
)

var (
	DefaultConfigPath     string
	DefaultKnownHostsPath string
	DefaultPrivateKeyPath string
)

var (
	Path   string
	Config ConfigData
)

type ConfigData struct {
	DeviceType     dsl.ClientType
	Host           string
	User           string
	PrivateKeyPath string
	KnownHostsPath string
	Options        map[string]string
}

func Load(path string) error {
	Path = path

	Config = ConfigData{
		PrivateKeyPath: DefaultPrivateKeyPath,
		KnownHostsPath: DefaultKnownHostsPath,
		Options:        make(map[string]string),
	}

	_, err := toml.DecodeFile(path, &Config)
	if path == DefaultConfigPath && errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return err
}

func Save() error {
	err := os.MkdirAll(filepath.Dir(Path), os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := toml.NewEncoder(file)

	if Config.DeviceType != "" {
		err = enc.Encode(map[string]string{"DeviceType": string(Config.DeviceType)})
		if err != nil {
			return err
		}
	}

	if Config.Host != "" {
		err = enc.Encode(map[string]string{"Host": Config.Host})
		if err != nil {
			return err
		}
	}

	if Config.User != "" {
		err = enc.Encode(map[string]string{"User": Config.User})
		if err != nil {
			return err
		}
	}

	if Config.PrivateKeyPath != DefaultPrivateKeyPath {
		err = enc.Encode(map[string]string{"PrivateKeyPath": Config.PrivateKeyPath})
		if err != nil {
			return err
		}
	}

	if Config.KnownHostsPath != DefaultKnownHostsPath {
		err = enc.Encode(map[string]string{"KnownHostsPath": Config.KnownHostsPath})
		if err != nil {
			return err
		}
	}

	err = enc.Encode(map[string]map[string]string{"Options": Config.Options})
	if err != nil {
		return err
	}

	return nil
}

func Validate() error {
	if !Config.DeviceType.IsValid() {
		return errors.New("invalid or missing device type")
	}
	clientDesc := Config.DeviceType.ClientDesc()

	if Config.Host == "" {
		return errors.New("no hostname specified")
	}

	if clientDesc.RequiresUser == dsl.TristateNo && Config.User != "" {
		return errors.New("username specified, but not required for device")
	} else if clientDesc.RequiresUser == dsl.TristateYes && Config.User == "" {
		return errors.New("no username specified")
	}

	for optionKey := range Config.Options {
		valid := false
		for option := range clientDesc.Options {
			if optionKey == option {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("invalid device-specific option: " + optionKey)
		}
	}

	return nil
}

func loadKnownHosts(file string) (string, error) {
	if file == "" {
		return "", nil
	}

	data, err := os.ReadFile(file)
	if err != nil && file != DefaultKnownHostsPath {
		return "", err
	}

	return string(data), nil
}

func loadPrivateKeys(file string) ([]string, error) {
	if file == "" {
		return nil, nil
	}

	if file[len(file)-1] == filepath.Separator {
		var keys []string

		keyFileNames := []string{"id_ed25519", "id_rsa", "id_ecdsa"}
		for _, name := range keyFileNames {
			data, err := os.ReadFile(file + name)
			if err != nil && file != DefaultPrivateKeyPath {
				return []string{}, err
			}

			if err == nil {
				keys = append(keys, string(data))
			}
		}

		return keys, nil
	}

	data, err := os.ReadFile(file)
	if err != nil && file != DefaultPrivateKeyPath {
		return []string{}, err
	}

	return []string{string(data)}, nil
}

func ClientConfig() (dsl.Config, error) {
	clientDesc := Config.DeviceType.ClientDesc()

	var knownHosts string
	if clientDesc.RequiresKnownHosts {
		if Config.KnownHostsPath == "IGNORE" {
			knownHosts = "IGNORE"
			fmt.Println("WARNING: Host key validation disabled!")
		} else {
			var err error
			knownHosts, err = loadKnownHosts(Config.KnownHostsPath)
			if err != nil {
				return dsl.Config{}, fmt.Errorf("failed to load known hosts file: %w", err)
			}
		}
	}

	var privateKeysCallback dsl.PrivateKeysCallback
	if clientDesc.SupportedAuthTypes&dsl.AuthTypePrivateKeys != 0 {
		privateKeysCallback.Keys = func() ([]string, error) {
			keys, err := loadPrivateKeys(Config.PrivateKeyPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load private key file: %w", err)
			}
			return keys, nil
		}
	}

	clientConfig := dsl.Config{
		Type:            Config.DeviceType,
		Host:            Config.Host,
		User:            Config.User,
		AuthPrivateKeys: privateKeysCallback,
		KnownHosts:      knownHosts,
		Options:         Config.Options,
	}

	return clientConfig, nil
}

func init() {
	DefaultPrivateKeyPath = filepath.Join(xdg.Home, ".ssh") + string(filepath.Separator)
	DefaultKnownHostsPath = filepath.Join(xdg.Home, ".ssh", "known_hosts")

	DefaultConfigPath = filepath.Join(xdg.ConfigHome, "3e8.eu-go-dsl", "config.toml")
}
