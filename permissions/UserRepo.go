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

package permissions

import (
	"gopkg.in/yaml.v3"
	"io"
)

type UserRepo interface {
	FindUser(subject string) *User
}

type Repository struct {
	Users []User
}

func (r *Repository) FindUser(subject string) *User {
	for _, u := range r.Users {
		if u.Subject == subject || u.Subject == Wildcard {
			userCopy := u
			userCopy.Subject = subject
			return &userCopy
		}
	}

	// Empty permission set for this user
	return &User{
		Subject: subject,
	}
}

func (r *Repository) LoadPermissions(config io.Reader) error {
	decoder := yaml.NewDecoder(config)
	return decoder.Decode(&r.Users)
}
