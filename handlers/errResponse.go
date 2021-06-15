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

package handlers

import (
	"encoding/json"
	"github.com/D-Haven/fact-totem/permissions"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Status  int
	Message string
}

func createError(err error) ErrorResponse {
	r := ErrorResponse{
		Message: err.Error(),
	}

	switch err.(type) {
	case NotFound:
		r.Status = http.StatusNotFound
	case Deleted:
		r.Status = http.StatusGone
	case RecordExists:
		r.Status = http.StatusConflict
	case permissions.NotAuthorized:
		r.Status = http.StatusUnauthorized
	case Unprocessed:
		r.Status = http.StatusNotFound
	default:
		r.Status = http.StatusConflict
	}

	return r
}

func (e ErrorResponse) write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	content, err := json.Marshal(&e)
	if err != nil {
		log.Printf("error in error handler: %s", err)
		return
	}

	_, err = w.Write(content)
	if err != nil {
		log.Printf("error in error handler: %s", err)
		return
	}
}
