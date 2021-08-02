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

package webapi

import (
	"github.com/D-Haven/fact-totem/permissions"
	"net/http"
)

type AuthenticatedHandler func(w http.ResponseWriter, r *http.Request, user *permissions.User)

type AuthHandler struct {
	UserConfig permissions.Config
	Handler    AuthenticatedHandler
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	validator := h.UserConfig.Validator()

	token, err := validator.ValidToken(r.Header.Get("Authorization"))
	if err != nil {
		createError(permissions.NotAuthorized{Cause: err}).write(w)
		return
	}

	user := h.UserConfig.Repository.FindUser(token)

	h.Handler(w, r, user)
}
