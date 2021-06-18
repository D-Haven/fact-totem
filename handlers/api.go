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
	"encoding/json"
	"fmt"
	"github.com/D-Haven/fact-totem/eventstore"
	"github.com/D-Haven/fact-totem/permissions"
	"net/http"
	"regexp"
)

type Call struct {
	Action  Action
	Matcher *regexp.Regexp
}

type FactApi struct {
	EventStore eventstore.EventStore
}

func NewApi(path string) (*FactApi, error) {
	api := FactApi{}
	api.EventStore = eventstore.FileStore(path)
	api.EventStore.Register(map[string]interface{}{})
	api.EventStore.Register([]interface{}{})

	return &api, nil
}

func (api *FactApi) Handle(w http.ResponseWriter, r *http.Request, user *permissions.User) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Allow", "OPTIONS, POST")
		w.Header().Set("Access-Control-Request-Method", http.MethodPost)
		w.Header().Set("Access-Control-Request-Headers", "Authorization, Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if user == nil {
		createError(permissions.NotAuthorized{}).write(w)
		return
	}

	if r.Method != http.MethodPost {
		resp := ErrorResponse{
			Status:  http.StatusMethodNotAllowed,
			Message: fmt.Sprintf("unsupported HTTP method: %s", r.Method),
		}

		resp.write(w)
	}

	var req Request
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&req)
	if err != nil {
		createError(BadRequest{Cause: err}).write(w)
		return
	}

	switch req.Action {
	case Append:
		tail, err := api.Append(user, req.Aggregate, req.Key, req.Content)
		if err != nil {
			createError(err).write(w)
			return
		}
		w.Header().Set("Location", "/")
		send(w, http.StatusCreated, tail)
	case Read:
		read, err := api.Read(user, req.Aggregate, req.Key, req.Origin, req.PageSize)
		if err != nil {
			createError(err).write(w)
			return
		}
		send(w, http.StatusOK, read)
	case Tail:
		tail, err := api.Tail(user, req.Aggregate, req.Key)
		if err != nil {
			createError(err).write(w)
			return
		}
		send(w, http.StatusOK, tail)
	case Scan:
		scan, err := api.Scan(user, req.Aggregate, req.PageSize)
		if err != nil {
			createError(err).write(w)
			return
		}
		send(w, http.StatusOK, scan)
	}
}

func (api *FactApi) Append(user *permissions.User, agg string, key string, content interface{}) (*TailResponse, error) {
	err := user.CheckPermission(permissions.Append, agg)
	if err != nil {
		return nil, err
	}

	// Aggregate is handled by user permissions (empty aggregate is always denied)

	if len(key) == 0 {
		return nil, BadRequest{Element: "key"}
	}
	if content == nil {
		return nil, BadRequest{Element: "content"}
	}

	tail, err := api.EventStore.Append(agg, key, content)
	if err != nil {
		return nil, err
	}

	resp := TailResponse{
		Aggregate: agg,
		Key:       key,
		Data:      tail.Record,
		Total:     tail.Total,
	}
	return &resp, nil
}

func (api *FactApi) Read(user *permissions.User, aggregate string, key string, origin string, size int) (*ReadResponse, error) {
	err := user.CheckPermission(permissions.Read, aggregate)
	if err != nil {
		return nil, err
	}

	// Aggregate is handled by user permissions (empty aggregate is always denied)

	if len(key) == 0 {
		return nil, BadRequest{Element: "key"}
	}

	records, err := api.EventStore.Read(aggregate, key, origin, size)
	if err != nil {
		return nil, err
	}

	resp := ReadResponse{
		Aggregate: aggregate,
		Key:       key,
		Data:      records.List,
		Total:     records.Total,
		PageSize:  records.PageSize,
	}

	return &resp, nil
}

func (api *FactApi) Tail(user *permissions.User, aggregate string, key string) (*TailResponse, error) {
	err := user.CheckPermission(permissions.Read, aggregate)
	if err != nil {
		return nil, err
	}

	// Aggregate is handled by user permissions (empty aggregate is always denied)

	if len(key) == 0 {
		return nil, BadRequest{Element: "key"}
	}

	tail, err := api.EventStore.Tail(aggregate, key)
	if err != nil {
		return nil, err
	}

	resp := TailResponse{
		Aggregate: aggregate,
		Key:       key,
		Data:      tail.Record,
		Total:     tail.Total,
	}
	return &resp, nil
}

func (api *FactApi) Scan(user *permissions.User, aggregate string, pageSize int) (*ScanResponse, error) {
	err := user.CheckPermission(permissions.Scan, aggregate)
	if err != nil {
		return nil, err
	}

	panic("Scan not implemented")
}

func send(w http.ResponseWriter, httpStatus int, object interface{}) {
	if httpStatus == http.StatusNoContent {
		w.WriteHeader(httpStatus)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	err := json.NewEncoder(w).Encode(object)
	if err != nil {
		createError(err).write(w)
		return
	}
}
