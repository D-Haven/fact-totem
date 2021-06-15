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

func TestUser_CheckPermission_CanRead(t *testing.T) {
	u := User{
		Subject: "foo",
		Read:    []string{"bar"},
		Append:  []string{},
		Scan:    []string{},
	}

	if err := u.CheckPermission(Read, "bar"); err != nil {
		t.Errorf("Expected success but received error: %s", err)
	}

	if u.CheckPermission(Read, "fubar") == nil {
		t.Errorf("Expected NotAuthorized but was allowed")
	}

	if u.CheckPermission(Append, "bar") == nil {
		t.Errorf("Expected NotAuthorized but was allowed")
	}

	if u.CheckPermission(Scan, "bar") == nil {
		t.Errorf("Expected NotAuthorized but was allowed")
	}
}

func TestUser_CheckPermission_CanAppend(t *testing.T) {
	u := User{
		Subject: "baz",
		Read:    []string{},
		Append:  []string{"bar"},
		Scan:    []string{},
	}

	if err := u.CheckPermission(Append, "bar"); err != nil {
		t.Errorf("Expected success but received error: %s", err)
	}

	if u.CheckPermission(Append, "fubar") == nil {
		t.Errorf("Expected NotAuthorized but was allowed")
	}

	if u.CheckPermission(Read, "bar") == nil {
		t.Errorf("Expected NotAuthorized but was allowed")
	}

	if u.CheckPermission(Scan, "bar") == nil {
		t.Errorf("Expected NotAuthorized but was allowed")
	}
}

func TestUser_CheckPermission_CanScan(t *testing.T) {
	u := User{
		Subject: "baz",
		Read:    []string{},
		Append:  []string{},
		Scan:    []string{"bar"},
	}

	if err := u.CheckPermission(Scan, "bar"); err != nil {
		t.Errorf("Expected success but received error: %s", err)
	}

	if u.CheckPermission(Scan, "fubar") == nil {
		t.Errorf("Expected NotAuthorized but was allowed")
	}

	if u.CheckPermission(Read, "bar") == nil {
		t.Errorf("Expected NotAuthorized but was allowed")
	}

	if u.CheckPermission(Append, "bar") == nil {
		t.Errorf("Expected NotAuthorized but was allowed")
	}
}
