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
	"testing"
)

func TestRepository_LoadPermissions(t *testing.T) {
	repo := Repository{}

	const userStream = `
    - subject: system:serviceaccount:cibob-dev:vault
      read: [project,repository]
      append: [project]
      scan: []`

	err := repo.LoadPermissions(strings.NewReader(userStream))
	if err != nil {
		t.Error(err)
	}

	u := repo.FindUser("system:serviceaccount:cibob-dev:vault")
	if u.CheckPermission(Append, "project") != nil {
		t.Error("Expected to be able to append to the project aggregate")
	}
	if u.CheckPermission(Read, "project") != nil {
		t.Error("Expected to be able to read the project aggregate")
	}
	if u.CheckPermission(Read, "repository") != nil {
		t.Error("Expected to be able to read the repository aggregate")
	}
}