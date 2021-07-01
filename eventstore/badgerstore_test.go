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

func TestBadgerEventStoreTail(t *testing.T) {
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

	add, err := store.Append(aggregate, key, content)
	if err != nil {
		t.Error(err)
	}

	tail, err := store.Tail(aggregate, key)
	if err != nil {
		t.Error(err)
		return
	}

	if add.Fact.Id != tail.Fact.Id {
		t.Errorf("expected tail '%s' but received tail '%s'", add.Fact.Id, tail.Fact.Id)
	}
}

func TestBadgerEventStoreAppend(t *testing.T) {
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

	tail, err := store.Append(aggregate, key, content)
	if err != nil {
		t.Error(err)
	}

	evtId := tail.Fact.Id

	if len(evtId.String()) == 0 {
		t.Error("expected the eventId to be returned")
	}

	results, err := store.Read(aggregate, key, "", -1)
	if err != nil {
		t.Error(err)
		return
	}

	lastEvt := lastEvent(results).Id

	if evtId != lastEvt {
		t.Error("expected same event to be returned")
	}

	verifyListLength(t, results, 1)

	if !reflect.DeepEqual(results.List[0].Content, content) {
		t.Errorf("Expected origional content \"%v\", but received \"%v\"", content, results.List[0].Content)
	}
}

func TestBadgerEventStoreAppendWithMultipleFacts(t *testing.T) {
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

	results1, err := store.Read(aggregate, key1, "", -1)
	if err != nil {
		t.Error(err)
	}

	if results1.Total != 3 {
		t.Errorf("Incorrect number of events: %d", results1.Total)
	}

	results2, err := store.Read(aggregate, key2, "", -1)
	if err != nil {
		t.Error(err)
	}

	if results2.Total != 9 {
		t.Errorf("Incorrect number of events: %d", results2.Total)
	}
}

func TestBadgerEventStoreReadFrom(t *testing.T) {
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
		add, err := store.Append(aggregate, key, Test{Value: i})
		if err != nil {
			t.Error(err)
		}
		lastEvt = add.Fact.Id.String()
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
		return
	}

	if tail.Fact.Id.String() <= lastEvt {
		t.Errorf("There should be events after the last event we captured.  Captured: %s Tail: %s", lastEvt, tail.Fact.Id)
	}

	results, err := store.Read(aggregate, key, lastEvt, -1)
	if err != nil {
		t.Error(err)
		return
	}

	if results.Total != 10 {
		t.Errorf("expected %d grand total events in the store, but there are only %d", 10, results.Total)
	}

	verifyListLength(t, results, 5)

	lastEvt = lastEvent(results).Id.String()

	if tail.Fact.Id.String() != lastEvt {
		t.Errorf("after reading from the event, expected to be caught up.  Tail %s, Last Event %s", tail.Fact.Id, lastEvt)
	}
}

func TestBadgerEventStoreScanAggregate(t *testing.T) {
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

	keys, err := store.Scan(aggregate)
	if err != nil {
		t.Error(err)
	}

	if int(keys.Total) != targetKeys {
		t.Errorf("expected %d keys, received %d", targetKeys, keys.Total)
	}

	for i, key := range keys.List {
		expected := string('a' + byte(i))
		if key != expected {
			t.Errorf("unexpected key: expected %s, received %s", expected, key)
		}
	}
}

func lastEvent(results *RecordList) Fact {
	return results.List[len(results.List)-1]
}

func verifyListLength(t *testing.T, results *RecordList, expectedSize int) {
	if len(results.List) != expectedSize {
		t.Errorf("expected %d events, received %d events", expectedSize, len(results.List))
	}
}
