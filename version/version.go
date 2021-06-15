/*
 * Copyright (c) 2021. D-Haven
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

package version

import (
	"encoding/json"
	"io"
)

var (
	// Release is the semantic version of the server
	Release = "unset"
	// BuildTime is when the server was built
	BuildTime = "unset"
	// Commit is the last commit hash when the server was built
	Commit = "unset"
)

type version struct {
	Release   string `json:"release"`
	BuildTime string `json:"build-time"`
	Commit    string `json:"commit"`
}

func Print(writer io.Writer) error {
	ver := version{
		Release:   Release,
		BuildTime: BuildTime,
		Commit:    Commit,
	}

	versionJson, err := json.MarshalIndent(&ver, "", "  ")
	if err != nil {
		return err
	}

	_, err = writer.Write(versionJson)
	if err != nil {
		return err
	}

	_, err = io.WriteString(writer, "\n")
	return err
}
