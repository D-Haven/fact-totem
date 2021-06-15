/*
 * Copyright (c) 2021.  D-Haven.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"github.com/D-Haven/fact-totem/handlers"
	"github.com/D-Haven/fact-totem/permissions"
	"github.com/lestrrat-go/jwx/jwa"
	"log"
	"math/rand"
	"os"
)

func ConfigureJwt(config *Config) {
	// Need to make this k8s service account compliant
	handlers.JwtSigType = jwa.HS512

	content, err := os.ReadFile(config.JwtKeyPath)
	if os.IsNotExist(err) {
		content = make([]byte, 4096)
		rand.Read(content)
		err := os.WriteFile(config.JwtKeyPath, content, 0640)
		if err != nil {
			log.Panicf("Could not create key jwt key file: %s", err)
		}
	} else if err != nil {
		log.Panicf("Unexpected error reading the key: %s", err)
	}

	handlers.JwtKey = content

	repo := permissions.Repository{}
	file, err := os.Open(config.PermissionsPath)
	if err != nil {
		log.Panicf("Could not open the permissions file: %s", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing config file: %s", err)
		}
	}()

	err = repo.LoadPermissions(file)
	if err != nil {
		log.Panicf("Could not read the permissions file: %s", err)
	}

	handlers.UserRepo = &repo
}
