/*
 * Copyright (c) 2021.   D-Haven.org
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package webapi

import (
	"crypto/rand"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"net/http"
	"os"
	"time"
)

type Validator interface {
	RetrieveToken(r *http.Request) (jwt.Token, error)
	IsValid(token jwt.Token) error
}

type JwtConfig struct {
	// SignatureType is the signature type we are expecting
	SignatureType jwa.SignatureAlgorithm `yaml:"SignatureType,omitempty"`
	// KeyPath is the path to the key file(s) (if using JWS)
	KeyPath string `yaml:"keyPath,omitempty"`
	// Issuer is the authority that issued the JWT
	Issuer string `yaml:"issuer,omitempty"`
	// Audience is the expected audience (i.e. "Fact-Totem")
	Audience string `yaml:"audience,omitempty"`
	// AcceptableSkew is the amount of time difference acceptable when testing times
	AcceptableSkew time.Duration `yaml:"acceptableSkew,omitempty"`
}

func (c *JwtConfig) Validator() (Validator, error) {
	v := jwtValidator{
		config: *c,
	}

	if len(c.SignatureType) == 0 || len(c.KeyPath) == 0 {
		return &v, nil
	}

	content, err := os.ReadFile(c.KeyPath)
	if os.IsNotExist(err) {
		content = make([]byte, 4096)
		_, err := rand.Read(content)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(c.KeyPath, content, 0640)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	v.jwtKey = content

	return &v, nil
}

type jwtValidator struct {
	// config the parent configuration
	config JwtConfig
	// jwtKey is the key content in bytes
	jwtKey []byte
}

func (v *jwtValidator) RetrieveToken(r *http.Request) (jwt.Token, error) {
	var options []jwt.ParseOption

	if len(v.config.SignatureType) != 0 {
		options = append(options, jwt.WithVerify(v.config.SignatureType, v.jwtKey))
	}

	return jwt.ParseRequest(r, options...)
}

func (v *jwtValidator) IsValid(token jwt.Token) error {
	var options []jwt.ValidateOption

	if len(v.config.Issuer) != 0 {
		options = append(options, jwt.WithIssuer(v.config.Issuer))
	}

	if len(v.config.Audience) != 0 {
		options = append(options, jwt.WithAudience(v.config.Audience))
	}

	if v.config.AcceptableSkew.Nanoseconds() > 0 {
		options = append(options, jwt.WithAcceptableSkew(v.config.AcceptableSkew))
	}

	return jwt.Validate(token, options...)
}
