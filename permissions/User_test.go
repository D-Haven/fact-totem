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

import "testing"

func TestUserCheckPermissionAllowWildcard(t *testing.T) {
	u := User{
		Subject: "test",
		Read:    []string{"*"},
		Append:  []string{"*"},
		Scan:    []string{"*"},
	}

	verifyPermitted(t, u, Read, "foo")
	verifyPermitted(t, u, Append, "bar")
	verifyPermitted(t, u, Scan, "baz")
}

func TestUserCheckPermissionNeverAllowEmptyAggregate(t *testing.T) {
	u := User{
		Subject: "test",
		Read:    []string{""},
		Append:  []string{""},
		Scan:    []string{""},
	}

	verifyDenied(t, u, Read, "")
	verifyDenied(t, u, Append, "")
	verifyDenied(t, u, Scan, "")
}

func TestUserCheckPermissionCanRead(t *testing.T) {
	u := User{
		Subject: "foo",
		Read:    []string{"bar"},
		Append:  []string{},
		Scan:    []string{},
	}

	verifyPermitted(t, u, Read, "bar")
	verifyDenied(t, u, Read, "fubar")
	verifyDenied(t, u, Append, "bar")
	verifyDenied(t, u, Scan, "bar")
}

func TestUserCheckPermissionCanAppend(t *testing.T) {
	u := User{
		Subject: "baz",
		Read:    []string{},
		Append:  []string{"bar"},
		Scan:    []string{},
	}

	verifyPermitted(t, u, Append, "bar")
	verifyDenied(t, u, Read, "bar")
	verifyDenied(t, u, Append, "fubar")
	verifyDenied(t, u, Scan, "bar")
}

func TestUserCheckPermissionCanScan(t *testing.T) {
	u := User{
		Subject: "baz",
		Read:    []string{},
		Append:  []string{},
		Scan:    []string{"bar"},
	}

	verifyPermitted(t, u, Scan, "bar")
	verifyDenied(t, u, Read, "bar")
	verifyDenied(t, u, Append, "bar")
	verifyDenied(t, u, Scan, "fubar")
}

func verifyPermitted(t *testing.T, u User, perm string, aggregate string) {
	if err := u.CheckPermission(perm, aggregate); err != nil {
		t.Errorf("expected %s:%s permission but was %s", perm, aggregate, err)
	}
}

func verifyDenied(t *testing.T, u User, perm string, aggregate string) {
	if u.CheckPermission(perm, aggregate) == nil {
		t.Errorf("expected NotAuthorized but was allowed")
	}
}
