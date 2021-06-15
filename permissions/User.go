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
	"strings"
)

const (
	Read   = "read"
	Append = "append"
	Scan   = "scan"
)

type User struct {
	Subject string
	Roles   []string
}

func (u User) CheckPermission(permission string, aggregate string) error {
	for _, r := range u.Roles {
		p := strings.Split(r, ":")
		if len(p) > 1 && strings.EqualFold(p[0], permission) {
			for _, agg := range strings.Split(p[1], ",") {
				if agg == aggregate {
					return nil
				}
			}
		}
	}

	return NotAuthorized{}
}
