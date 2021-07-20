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

package k8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	tokenReviewKind = "TokenReview"
	apiVersion      = "authentication.k8s.io/v1"
)

type TokenReview struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
	Spec       struct {
		Token string `json:"token"`
	} `json:"spec"`
	Status struct {
		Authenticated bool `json:"authenticated,omitempty"`
	} `json:"status,omitempty"`
}

func (c *Client) ValidateToken(token string) (*TokenReview, error) {
	trRequest := TokenReview{
		Kind:       tokenReviewKind,
		ApiVersion: apiVersion,
	}

	trRequest.Spec.Token = token
	trJson, err := json.Marshal(trRequest)
	if err != nil {
		return nil, err
	}

	resp, err := c.Post("authentication.k8s.io/v1/tokenreviews", bytes.NewBuffer(trJson))

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token review request failed: %d %s\n%s", resp.StatusCode, resp.Status, string(body))
	}

	trResponse := TokenReview{}
	err = json.Unmarshal(body, &trResponse)
	if err != nil {
		return nil, err
	}

	return &trResponse, nil
}
