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
	Fact  Fact
	Total uint
}

type RecordList struct {
	List     []Fact
	Total    uint
	PageSize int
}

type EntityList struct {
	List  []string
	Total uint
}

// EventStore provides an interface to store events for a topic, and retrieve them later.
type EventStore interface {
	// Register a type for (de)serialization, needed to store and reconstitute objects
	Register(t interface{})
	// Append append an event to the event store for the fact
	Append(aggregate string, entity string, content interface{}) (*Tail, error)
	// Tail gets the last event id
	Tail(aggregate string, entity string) (*Tail, error)
	// Read the events for an aggregate from the identified event id
	Read(aggregate string, entity string, originEventId string, maxCount int) (*RecordList, error)
	// Scan will list all keys in the aggregate (excluding individual events)
	Scan(aggregate string) (*EntityList, error)
	// Close the event store
	Close() error
}
