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

// Package eventstore handles the low level badger interactions.
package eventstore

type Tail struct {
	Record Record
	Total  uint
}

type RecordList struct {
	List     []Record
	Total    uint
	PageSize int
}

type KeyList struct {
	List     []string
	Total    uint
	PageSize int
}

// EventStore provides an interface to store events for a topic, and retrieve them later.
type EventStore interface {
	// Register a type for (de)serialization, needed to store and reconstitute objects
	Register(t interface{})
	// Append append an event to the event store for the fact
	Append(aggregate string, key string, content interface{}) (*Tail, error)
	// Tail gets the last event id
	Tail(aggregate string, key string) (*Tail, error)
	// Read the events for an aggregate from the identified event id
	Read(aggregate string, key string, originEventId string, maxCount int) (*RecordList, error)
	// ListKeysForAggregate will list all keys with the aggregate prefix
	ListKeysForAggregate(aggregate string) ([]string, error)
	// Close the event store
	Close() error
}
