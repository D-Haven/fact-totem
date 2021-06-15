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

package handlers

import (
	"github.com/D-Haven/fact-totem/permissions"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"net/http"
	"time"
)

var (
	JwtSigType jwa.SignatureAlgorithm
	JwtKey     []byte
)

type AuthenticatedHandler func(w http.ResponseWriter, r *http.Request, user *Token)

type Token struct {
	Version string   `json:"ver"`
	Subject string   `json:"sub"`
	Roles   []string `json:"rol"`
}

func (t *Token) User() *permissions.User {
	return &permissions.User{
		Subject: t.Subject,
		Roles:   t.Roles,
	}
}

func AuthHandler(authHandler AuthenticatedHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _ := jwt.ParseRequest(r, jwt.WithVerify(JwtSigType, JwtKey))
		var user *Token

		if token != nil {
			err := jwt.Validate(token,
				jwt.WithAudience("pmbob"),
				jwt.WithIssuer("hrbob"),
				jwt.WithAcceptableSkew(2*time.Second))
			if err != nil {
				createError(permissions.NotAuthorized{Cause: err}).write(w)
				return
			}

			/*
				iat, iatOk := token.Get("iat")
				exp, expOk := token.Get("exp")

				if !iatOk || !expOk || exp.(time.Time).Sub(iat.(time.Time)) > 10*time.Minute {
					createError(NotAuthorized{
						Cause: errors.New("invalid token spec"),
					}).write(w)
					return
				}
			*/

			user = &Token{
				Subject: token.Subject(),
				Version: token.PrivateClaims()["ver"].(string),
				Roles:   token.PrivateClaims()["rol"].([]string),
			}
		}

		authHandler(w, r, user)
	})
}
