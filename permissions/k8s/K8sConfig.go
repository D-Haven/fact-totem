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

package k8s

import (
	"fmt"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"strings"
)

type K8sConfig struct {
	Enabled bool   `yaml:"enabled"`
	client  Client `yaml:"-"`
}

func (v *K8sConfig) ValidToken(token string) (jwt.Token, error) {
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	jot, err := jwt.Parse([]byte(token))
	if err != nil {
		return nil, err
	}

	tr, err := v.client.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	if !tr.Status.Authenticated {
		return nil, fmt.Errorf("k8s service account '%s' is not authenticated", jot.Subject())
	}

	return jot, nil
}
