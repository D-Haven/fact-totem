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

import "github.com/D-Haven/fact-totem/eventstore"

type Request struct {
	Action    Action      `json:"action"`
	Aggregate string      `json:"aggregate"`
	Entity    string      `json:"entity,omitempty"`
	Content   interface{} `json:"content,omitempty"`
	Origin    string      `json:"origin,omitempty"`
	PageSize  int         `json:"page-size,omitempty"`
}

type TailResponse struct {
	Aggregate string            `json:"aggregate"`
	Entity    string            `json:"entity,omitempty"`
	Data      eventstore.Record `json:"data"`
	Total     uint              `json:"total"`
}

type ReadResponse struct {
	Aggregate string              `json:"aggregate"`
	Entity    string              `json:"entity"`
	Data      []eventstore.Record `json:"data"`
	Total     uint                `json:"total"`
	PageSize  int                 `json:"page-size"`
}

type ScanResponse struct {
	Aggregate string   `json:"aggregate"`
	Entities  []string `json:"entities"`
	Total     uint     `json:"total"`
}
