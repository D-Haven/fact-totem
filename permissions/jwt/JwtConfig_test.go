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

package jwt

import (
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"testing"
	"time"
)

func TestJwtConfigValidTokenWithAudience(t *testing.T) {
	config := JwtConfig{
		KeyAlgorithm: jwa.HS512,
		Audience:     "mytest",
		KeyPath:      "../jwt.key",
	}
	err := config.loadKey()
	if err != nil {
		t.Error(err)
		return
	}

	good := jwt.New()
	_ = good.Set(jwt.SubjectKey, "test")
	_ = good.Set(jwt.AudienceKey, config.Audience)

	bad := jwt.New()
	_ = bad.Set(jwt.SubjectKey, "test")
	_ = bad.Set(jwt.AudienceKey, "not right")

	gb, err := jwt.Sign(good, jwt.WithKey(config.KeyAlgorithm, config.jwtKey))
	if err != nil {
		t.Error(err)
		return
	}
	bb, err := jwt.Sign(bad, jwt.WithKey(config.KeyAlgorithm, config.jwtKey))
	if err != nil {
		t.Error(err)
		return
	}

	tok, err := config.ValidToken(string(gb))
	if err != nil {
		t.Error(err)
	}
	if tok.Subject() != good.Subject() {
		t.Errorf("Good token mismatch on validation")
	}

	_, err = config.ValidToken(string(bb))
	if err == nil {
		t.Error("Expected error, but did not receive it.")
	}
}

func TestJwtConfigValidTokenWithValidTimeSpan(t *testing.T) {
	config := JwtConfig{
		KeyAlgorithm:   jwa.HS512,
		KeyPath:        "../jwt.key",
		MaxValidWindow: 10 * time.Minute,
		AcceptableSkew: 50 * time.Millisecond,
	}
	err := config.loadKey()
	if err != nil {
		t.Error(err)
		return
	}

	good := jwt.New()
	_ = good.Set(jwt.SubjectKey, "test")
	_ = good.Set(jwt.IssuedAtKey, time.Now())
	_ = good.Set(jwt.ExpirationKey, time.Now().Add(10*time.Minute))

	gb, err := jwt.Sign(good, jwt.WithKey(config.KeyAlgorithm, config.jwtKey))
	if err != nil {
		t.Error(err)
		return
	}

	tok, err := config.ValidToken(string(gb))
	if err != nil {
		t.Error(err)
	}
	if tok.Subject() != good.Subject() {
		t.Errorf("Good token mismatch on validation")
	}
}
