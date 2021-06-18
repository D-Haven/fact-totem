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

package handlers

import (
	"bytes"
	"encoding/json"
)

type Action int

const (
	Append Action = iota
	Read
	Tail
	Scan
)

func (a Action) String() string {
	return toString[a]
}

var toString = map[Action]string{
	Append: "Append",
	Read:   "Read",
	Tail:   "Tail",
	Scan:   "Scan",
}

var toId = map[string]Action{
	"Append": Append,
	"Read":   Read,
	"Tail":   Tail,
	"Scan":   Scan,
}

// MarshalJSON marshals the enum as a quoted json string
func (a Action) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[a])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (a *Action) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Timeout' in this case.
	*a = toId[j]
	return nil
}
