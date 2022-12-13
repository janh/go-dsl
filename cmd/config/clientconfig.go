// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"3e8.eu/go/dsl"
)

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

	var passwordCallback dsl.PasswordCallback
	if clientDesc.SupportedAuthTypes&dsl.AuthTypePassword != 0 {
		if Secrets.Password != "" {
			passwordCallback = dsl.Password(Secrets.Password)
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
		if Secrets.PrivateKeyPassphrase != "" {
			privateKeysCallback.Passphrase = func(string) (string, error) {
				return Secrets.PrivateKeyPassphrase, nil
			}
		}
	}

	var encryptionPassphraseCallback dsl.EncryptionPassphraseCallback
	if clientDesc.SupportsEncryptionPassphrase {
		if Secrets.EncryptionPassphrase != "" {
			encryptionPassphraseCallback = dsl.EncryptionPassphrase(Secrets.EncryptionPassphrase)
		}
	}

	clientConfig := dsl.Config{
		Type:                 Config.DeviceType,
		Host:                 Config.Host,
		User:                 Config.User,
		AuthPassword:         passwordCallback,
		AuthPrivateKeys:      privateKeysCallback,
		EncryptionPassphrase: encryptionPassphraseCallback,
		KnownHosts:           knownHosts,
		Options:              Config.Options,
	}

	return clientConfig, nil
}
