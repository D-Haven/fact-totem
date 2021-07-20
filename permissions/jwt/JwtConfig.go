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

package jwt

import (
	"crypto/rand"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"os"
	"strings"
	"time"
)

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
	// Maximum duration between the jwt issue time and expiration time, implicitly requires both values
	MaxValidWindow time.Duration `yaml:"maxValidWindow,omitempty"`
	jwtKey         []byte        `yaml:"-"`
}

func (c *JwtConfig) loadKey() error {
	if len(c.SignatureType) == 0 || len(c.KeyPath) == 0 {
		return nil
	}

	content, err := os.ReadFile(c.KeyPath)
	if os.IsNotExist(err) {
		content = make([]byte, 4096)
		_, err := rand.Read(content)
		if err != nil {
			return err
		}

		err = os.WriteFile(c.KeyPath, content, 0640)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	c.jwtKey = content

	return nil
}

func (c *JwtConfig) ValidToken(token string) (jwt.Token, error) {
	if (len(c.SignatureType) > 0 || len(c.KeyPath) > 0) && len(c.jwtKey) == 0 {
		err := c.loadKey()
		if err != nil {
			return nil, err
		}
	}
	var parsOpt []jwt.ParseOption

	if len(c.SignatureType) != 0 {
		parsOpt = append(parsOpt, jwt.WithVerify(c.SignatureType, c.jwtKey))
	}

	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	jot, err := jwt.Parse([]byte(token), parsOpt...)
	if err != nil {
		return nil, err
	}

	var valOpt []jwt.ValidateOption

	if len(c.Issuer) != 0 {
		valOpt = append(valOpt, jwt.WithIssuer(c.Issuer))
	}

	if len(c.Audience) != 0 {
		valOpt = append(valOpt, jwt.WithAudience(c.Audience))
	}

	if c.AcceptableSkew.Nanoseconds() > 0 {
		valOpt = append(valOpt, jwt.WithAcceptableSkew(c.AcceptableSkew))
	}

	if c.MaxValidWindow.Nanoseconds() > 0 {
		valOpt = append(valOpt, jwt.WithRequiredClaim(jwt.IssuedAtKey))
		valOpt = append(valOpt, jwt.WithRequiredClaim(jwt.ExpirationKey))
		valOpt = append(valOpt, jwt.WithMaxDelta(c.MaxValidWindow, jwt.ExpirationKey, jwt.IssuedAtKey))
	}

	return jot, jwt.Validate(jot, valOpt...)
}
