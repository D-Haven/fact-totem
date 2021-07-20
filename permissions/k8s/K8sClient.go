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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	http      *http.Client
	caCert    []byte
	saToken   []byte
	namespace string
	rootUrl   string
}

func (c *Client) Post(api string, body io.Reader) (*http.Response, error) {
	if c.http == nil {
		err := c.Refresh()
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(http.MethodPost, strings.Join([]string{c.rootUrl, api}, "/"), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", string(c.saToken)))

	return c.http.Do(req)
}

func (c *Client) Refresh() error {
	host, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	if !ok {
		return fmt.Errorf("could not resolve %s", "KUBERNETES_SERVICE_HOST")
	}
	port, ok := os.LookupEnv("KUBERNETES_SERVICE_PORT_HTTPS")
	if !ok {
		return fmt.Errorf("could not resolve %s", "KUBERNETES_SERVICE_PORT_HTTPS")
	}
	c.rootUrl = fmt.Sprintf("https://%s:%s/apis", host, port)

	myToken, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return err
	}
	c.saToken = myToken

	caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return err
	}
	c.caCert = caCert
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(c.caCert)

	c.http = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	return nil
}
