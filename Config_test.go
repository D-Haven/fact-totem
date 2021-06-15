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
	"os"
	"strings"
	"testing"
)

func Check(t *testing.T, field string, expected string, actual string) {
	if expected != actual {
		t.Fatalf("%s expected %s but received %s", field, expected, actual)
	}
}

func TestMain(m *testing.M) {
	_ = os.Mkdir("./www", os.ModeTemporary)
	f, _ := os.Create("tls.crt")
	_ = f.Close()
	f, _ = os.Create("tls.key")
	_ = f.Close()

	code := m.Run()

	_ = os.RemoveAll("./www")
	_ = os.Remove("tls.crt")
	_ = os.Remove("tls.key")

	os.Exit(code)
}

func TestReadValidConfigFromYaml(t *testing.T) {
	content := `
jwtKeyPath: ./jwt.key
server:
  port: 8443
  tls:
    certificate: /.cert/tls.crt
    key: /.cert/tls.key`

	reader := strings.NewReader(content)

	config, err := ReadConfig(reader)
	if err != nil {
		t.Fatalf("Yaml read error: %s", err)
	}

	if config == nil {
		t.Fatal("No config was created")
	}

	Check(t, "server:port", "8443", config.Server.Port)
	Check(t, "jwtKeyPath:", "./jwt.key", config.JwtKeyPath)
	Check(t, "server:tls:certificate", "/.cert/tls.crt", config.Server.TLS.CertFile)
	Check(t, "server:tls:key", "/.cert/tls.key", config.Server.TLS.KeyFile)
}

func TestReadInvalidConfigFromYaml(t *testing.T) {
	content := "This is not YAML!!!"

	reader := strings.NewReader(content)

	config, err := ReadConfig(reader)
	if config != nil {
		t.Fatalf("An invalid config object was returned: %s", config)
	}

	if err == nil {
		t.Fatal("Invalid content was provided, but no error was produced")
	}
}

func TestValidateConfigIfKeySpecifiedCertIsRequired(t *testing.T) {
	config := &Config{}
	config.Server.TLS.KeyFile = "specified.pem"

	err := ValidateConfig(config)
	if err == nil {
		t.Fatal("Expected error because Server:TLS:KeyFile had a value but Server:TLS:CertFile did not")
	}
}

func TestValidateConfigIfCertSpecifiedKeyIsRequired(t *testing.T) {
	config := &Config{}
	config.Server.TLS.CertFile = "specified.pem"

	err := ValidateConfig(config)
	if err == nil {
		t.Fatal("Expected error because Server:TLS:CertFile had a value but Server:TLS:KeyFile did not")
	}
}

func TestValidateConfigWithBothCertAndKeyValidFiles(t *testing.T) {
	config := &Config{}
	config.Server.TLS.CertFile = "LICENSE"
	config.Server.TLS.KeyFile = "CONTRIBUTING.md"

	err := ValidateConfig(config)
	if err != nil {
		t.Fatalf("Should be a valid configuration since files exist: %s", err)
	}
}

func TestValidateOptionalFileIfExists(t *testing.T) {
	path := "./Logo.go"
	err := ValidateOptionalFile(path)

	if err != nil {
		t.Fatalf("./Logo.go is a valid file, but received error: %s", err)
	}
}

func TestValidateOptionalFileIfDirectory(t *testing.T) {
	path := "./handlers"
	err := ValidateOptionalFile(path)

	if err == nil {
		t.Fatalf("./handlers is a directory, but no error returned")
	}
}

func TestValidateOptionalFileIfNoExists(t *testing.T) {
	path := "/dev/null/nofile.config"
	err := ValidateOptionalFile(path)

	if err == nil {
		t.Fatalf("/dev/null/nofile.config should never exist, but no error returned")
	}
}

func TestGetValidatedConfigSuccess(t *testing.T) {
	config, err := GetValidatedConfig("./config.yaml")

	if err != nil {
		t.Fatal(err)
	}

	if config == nil {
		t.Fatal("No error was specified, but config was empty")
	}
}

func TestGetValidatedConfigFailure(t *testing.T) {
	config, err := GetValidatedConfig("/dev/null/config.yaml")

	if err != nil {
		t.Fatal(err)
	}

	if config == nil {
		t.Fatal("No error was specified, but config was empty")
	}

	Check(t, "JwtKeyPath", "./jwt.key", config.JwtKeyPath)
	Check(t, "Server:Port", "8080", config.Server.Port)
	Check(t, "Server:TLS:KeyFile", "", config.Server.TLS.KeyFile)
	Check(t, "Server:TLS:CertFile", "", config.Server.TLS.CertFile)
}
