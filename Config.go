/*
 * Copyright (c) 2021.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
)

type Config struct {
	// JWT Key path
	JwtKeyPath string `yaml:"jwtKeyPath"`
	// Project Db path
	ProjectDb string `yaml:"project-db"`

	Server struct {
		// Host is the server host name
		Host string `yaml:"host"`

		// Port is the local machine TCP Port to bind the HTTP Server to
		Port string `yaml:"port"`

		// TLS provides the TLS tuning configuration
		TLS struct {
			CertFile string `yaml:"certificate"`
			KeyFile  string `yaml:"key"`
		} `yaml:"tls"`
	} `yaml:"server"`
}

func GetValidatedConfig(configPath string) (*Config, error) {
	var config *Config
	file, err := os.Open(configPath)
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing config file: %s", err)
		}
	}()

	if err == nil {
		config, err = ReadConfig(file)
	}

	if err != nil || config == nil {
		log.Printf("Reverting to defaults... error with config file: %s", err)
		config = &Config{}

		config.Server.Port = "8080"
		config.JwtKeyPath = "./jwt.key"
		config.ProjectDb = "./tmp/project"
	}

	err = ValidateConfig(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func ReadConfig(reader io.Reader) (*Config, error) {
	config := &Config{}

	d := yaml.NewDecoder(reader)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func ValidateConfig(config *Config) error {
	tlsCertSpecified := len(config.Server.TLS.CertFile) > 0
	tlsKeySpecified := len(config.Server.TLS.KeyFile) > 0

	if !tlsCertSpecified && !tlsKeySpecified {
		return nil
	}

	if tlsCertSpecified && !tlsKeySpecified {
		return fmt.Errorf("TLS Certificate specified, but no key file was provided")
	}

	if tlsKeySpecified && !tlsCertSpecified {
		return fmt.Errorf("TLS Key specified, but no certificate was provided")
	}

	if err := ValidateOptionalFile(config.Server.TLS.KeyFile); err != nil {
		return err
	}

	if err := ValidateOptionalFile(config.Server.TLS.CertFile); err != nil {
		return err
	}

	return nil
}

func ValidateOptionalFile(path string) error {
	if len(path) > 0 {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return fmt.Errorf("'%s' is directory, not a file", path)
		}
	}

	return nil
}
