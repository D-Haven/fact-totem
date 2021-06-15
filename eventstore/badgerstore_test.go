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

package eventstore

import (
	"reflect"
	"testing"
)

type Test struct {
	Value int
}

func TestBadgerEventStore_Tail(t *testing.T) {
	store := MemoryStore()
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	store.Register(Test{})
	aggregate := "test"
	key := "1"
	content := Test{Value: 1}

	evtId, err := store.Append(aggregate, key, content)
	if err != nil {
		t.Error(err)
	}

	tailId, err := store.Tail(aggregate, key)
	if evtId != tailId {
		t.Errorf("expected tail '%s' but received tail '%s'", evtId, tailId)
	}
}

func TestBadgerEventStore_Append(t *testing.T) {
	store := MemoryStore()
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	store.Register(Test{})
	aggregate := "test"
	key := "1"
	content := Test{Value: 1}

	evtId, err := store.Append(aggregate, key, content)
	if err != nil {
		t.Error(err)
	}

	if len(evtId) == 0 {
		t.Error("expected the eventId to be returned")
	}

	results, lastEvt, err := store.Read(aggregate, key)
	if err != nil {
		t.Error(err)
		return
	}

	if evtId != lastEvt {
		t.Error("expected same event to be returned")
	}

	if len(results) != 1 {
		t.Errorf("Incorrect number of events: %d", len(results))
	}

	if !reflect.DeepEqual(results[0], content) {
		t.Errorf("Expected origional content \"%v\", but received \"%v\"", content, results[0])
	}
}

func TestBadgerEventStore_AppendWithMultipleFacts(t *testing.T) {
	store := MemoryStore()
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()

	store.Register(Test{})

	aggregate := "test"
	key1 := "1"
	key2 := "barly"

	for i := 0; i < 3; i++ {
		value := Test{Value: i}

		_, err := store.Append(aggregate, key1, value)
		if err != nil {
			t.Errorf("Failed append %s:", err)
			break
		}
	}

	for i := 0; i < 9; i++ {
		value := Test{Value: i}

		_, err := store.Append(aggregate, key2, value)
		if err != nil {
			t.Errorf("Failed append %s:", err)
			break
		}
	}

	results1, _, err := store.Read(aggregate, key1)
	if err != nil {
		t.Error(err)
	}

	if len(results1) != 3 {
		t.Errorf("Incorrect number of events: %d", len(results1))
	}

	results2, _, err := store.Read(aggregate, key2)
	if err != nil {
		t.Error(err)
	}

	if len(results2) != 9 {
		t.Errorf("Incorrect number of events: %d", len(results2))
	}
}

func TestBadgerEventStore_ReadFrom(t *testing.T) {
	store := MemoryStore()
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()

	store.Register(Test{})
	aggregate := "captain"
	key := "caveman"

	var lastEvt string
	var err error

	for i := 0; i < 5; i++ {
		lastEvt, err = store.Append(aggregate, key, Test{Value: i})
		if err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 5; i++ {
		_, err := store.Append(aggregate, key, Test{Value: i + 5})
		if err != nil {
			t.Error(err)
		}
	}

	tail, err := store.Tail(aggregate, key)
	if err != nil {
		t.Error(err)
	}

	if tail <= lastEvt {
		t.Errorf("There should be events after the last event we captured.  Captured: %s Tail: %s", lastEvt, tail)
	}

	results, lastEvt, err := store.ReadFrom(aggregate, key, lastEvt)
	if err != nil {
		t.Error(err)
	}

	if len(results) != 5 {
		t.Errorf("expected %d results, received %d", 5, len(results))
	}

	if tail != lastEvt {
		t.Errorf("after reading from the event, expected to be caught up.  Tail %s, Last Event %s", tail, lastEvt)
	}
}

func TestBadgerEventStore_ListKeysForAggregate(t *testing.T) {
	store := MemoryStore()
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()

	store.Register(Test{})
	aggregate := "barney"
	targetKeys := 10

	for i := 0; i < targetKeys; i++ {
		key := string('a' + byte(i))

		for k := 0; k < 5; k++ {
			test := Test{Value: k}

			_, err := store.Append(aggregate, key, test)
			if err != nil {
				t.Error(err)
			}
		}
	}

	keys, err := store.ListKeysForAggregate(aggregate)
	if err != nil {
		t.Error(err)
	}

	if len(keys) != targetKeys {
		t.Errorf("expected %d keys, received %d", targetKeys, len(keys))
	}

	for i, key := range keys {
		expected := string('a' + byte(i))
		if key != expected {
			t.Errorf("unexpected key: expected %s, received %s", expected, key)
		}
	}
}
