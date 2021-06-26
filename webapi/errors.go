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
	"fmt"
	"strings"
)

type Unprocessed struct{}

type RecordExists struct {
	Name string
}

type NotFound struct {
	Id string
}

type Deleted struct {
	Id string
}

type BadRequest struct {
	Element string
	Cause   error
}

func (u Unprocessed) Error() string {
	return "resource does not exist"
}

func (p RecordExists) Error() string {
	return fmt.Sprintf("record name already in use: '%s'", p.Name)
}

func (p NotFound) Error() string {
	return fmt.Sprintf("project not found: %s", p.Id)
}

func (p Deleted) Error() string {
	return fmt.Sprintf("project deleted: %s", p.Id)
}

func (br BadRequest) Error() string {
	b := strings.Builder{}
	b.WriteString("invalid request format")

	if len(br.Element) > 0 {
		b.WriteString(" '")
		b.WriteString(br.Element)
		b.WriteString("'")
	}

	if br.Cause != nil {
		b.WriteString(" ")
		b.WriteString(br.Cause.Error())
	}

	return b.String()
}
