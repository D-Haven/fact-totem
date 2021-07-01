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
	"bytes"
	"encoding/gob"
	"github.com/dgraph-io/badger/v3"
	"github.com/oklog/ulid/v2"
	"strings"
	"time"
)

const (
	separator   = "|"
	maxPageSize = 10000
)

type BadgerEventStore struct {
	RootDir                    string
	MemoryOnly                 bool
	EncryptionKey              []byte
	EncryptionRotationDuration time.Duration
	db                         *badger.DB
	generator                  IdGenerator
}

type Fact struct {
	Id        ulid.ULID
	Timestamp time.Time
	Content   interface{}
}

type AggregateStats struct {
	LastId ulid.ULID
	Total  uint
}

func MemoryStore() EventStore {
	return &BadgerEventStore{
		MemoryOnly: true,
		RootDir:    "",
		generator:  NewIdGenerator(),
	}
}

func FileStore(path string) EventStore {
	return &BadgerEventStore{
		RootDir:       path,
		MemoryOnly:    false,
		EncryptionKey: nil,
		generator:     NewIdGenerator(),
	}
}

func EncryptedFileStore(path string, key []byte, rotationDur time.Duration) EventStore {
	return &BadgerEventStore{
		RootDir:                    path,
		MemoryOnly:                 false,
		EncryptionKey:              key,
		EncryptionRotationDuration: rotationDur,
		generator:                  NewIdGenerator(),
	}
}

func (b *BadgerEventStore) Register(t interface{}) {
	gob.Register(t)
}

func (b *BadgerEventStore) readEntityStats(txn *badger.Txn, aggregate string, entity string) (*AggregateStats, error) {
	stats := AggregateStats{}
	aggKey := b.aggregateKey(aggregate, entity)

	item, err := txn.Get(aggKey)
	if err == nil {
		// If there is an error then this is a virgin aggregate so there is nothing to read
		err = item.Value(func(val []byte) error {
			dec := gob.NewDecoder(bytes.NewBuffer(val))
			return dec.Decode(&stats)
		})

		if err != nil {
			return nil, err
		}
	}

	return &stats, nil
}

func (b *BadgerEventStore) updateEntityStats(txn *badger.Txn, aggregate string, entity string, stats *AggregateStats) error {
	aggKey := b.aggregateKey(aggregate, entity)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(stats)
	if err != nil {
		return err
	}

	return txn.Set(aggKey, buf.Bytes())
}

func (b *BadgerEventStore) Append(aggregate string, entity string, content interface{}) (*Tail, error) {
	db, err := b.kvStore()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	tail := Tail{
		Fact: Fact{
			Id:        b.generator.NewId(now),
			Timestamp: now,
			Content:   content,
		},
	}

	if err = db.Update(func(txn *badger.Txn) error {
		stats, err := b.readEntityStats(txn, aggregate, entity)
		factKey := b.factKey(aggregate, entity, tail.Fact.Id)

		value, err := encodeFact(tail.Fact)
		if err != nil {
			return err
		}

		entry := badger.NewEntry(factKey, value)
		err = txn.SetEntry(entry)
		if err != nil {
			return err
		}

		stats.LastId = tail.Fact.Id
		stats.Total += 1
		tail.Total = stats.Total

		return b.updateEntityStats(txn, aggregate, entity, stats)
	}); err != nil {
		return nil, err
	}

	return &tail, nil
}

func (b *BadgerEventStore) Read(aggregate string, entity string, factId string, maxCount int) (*RecordList, error) {
	db, err := b.kvStore()
	if err != nil {
		return nil, err
	}

	var records = RecordList{
		PageSize: maxCount,
	}

	if records.PageSize < 1 || records.PageSize > maxPageSize {
		records.PageSize = maxPageSize
	}

	aggKey := []byte(strings.Join([]string{aggregate, entity}, separator))
	factKey := []byte(strings.Join([]string{aggregate, entity, factId}, separator))

	if err = db.View(func(txn *badger.Txn) error {
		stats, err := b.readEntityStats(txn, aggregate, entity)
		if err != nil {
			return err
		}

		records.Total = stats.Total

		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		// Walk all the events using the aggregate as a prefix
		for it.Seek(factKey); len(records.List) < records.PageSize && it.ValidForPrefix(aggKey); it.Next() {
			item := it.Item()

			record, err := decodeFact(item)
			if err != nil {
				return err
			}

			if record.Id.String() > factId {
				records.List = append(records.List, *record)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &records, nil
}

func (b *BadgerEventStore) Tail(aggregate string, entity string) (*Tail, error) {
	db, err := b.kvStore()
	if err != nil {
		return nil, err
	}

	tail := Tail{}
	err = db.View(func(txn *badger.Txn) error {
		stats, err := b.readEntityStats(txn, aggregate, entity)
		if err != nil {
			return err
		}

		tail.Total = stats.Total

		evtKey := b.factKey(aggregate, entity, stats.LastId)
		item, err := txn.Get(evtKey)
		if err != nil {
			return err
		}

		record, err := decodeFact(item)
		if err != nil {
			return err
		}

		tail.Fact = *record
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &tail, nil
}

func (b *BadgerEventStore) Scan(aggregate string) (*EntityList, error) {
	db, err := b.kvStore()
	if err != nil {
		return nil, err
	}

	keys := EntityList{}

	prefix := []byte(aggregate)

	if err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = prefix

		it := txn.NewIterator(opts)
		defer it.Close()

		var lastKey = ""
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			id := string(item.Key())
			parts := strings.Split(id, separator)
			key := parts[1]

			if len(lastKey) == 0 || lastKey != key {
				keys.List = append(keys.List, key)
				keys.Total += 1
				lastKey = key
			}
		}

		if len(lastKey) > 0 && (len(keys.List) == 0 || keys.List[len(keys.List)-1] != lastKey) {
			keys.List = append(keys.List, lastKey)
			keys.Total += 1
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &keys, nil
}

func (b *BadgerEventStore) Close() error {
	if b.db != nil {
		if err := b.db.Close(); err != nil {
			return err
		}
		b.db = nil
	}

	return nil
}

func (b *BadgerEventStore) kvStore() (*badger.DB, error) {
	if b.db != nil {
		return b.db, nil
	}

	opts := badger.DefaultOptions(b.RootDir).WithInMemory(b.MemoryOnly)

	if b.EncryptionKey != nil && len(b.EncryptionKey) >= 128 {
		opts = opts.WithEncryptionKey(b.EncryptionKey)
		opts = opts.WithEncryptionKeyRotationDuration(b.EncryptionRotationDuration)
		// May need to tune this.. data store shouldn't get too big
		opts = opts.WithIndexCacheSize(100 << 20) // 100 mb
	}

	b.Register(Fact{})
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	b.db = db
	return b.db, nil
}

func (b *BadgerEventStore) aggregateKey(aggregate string, entity string) []byte {
	return []byte(strings.Join([]string{aggregate, entity}, separator))
}

func (b *BadgerEventStore) factKey(aggregate string, entity string, factId ulid.ULID) []byte {
	return []byte(strings.Join([]string{aggregate, entity, factId.String()}, separator))
}

func encodeFact(fact Fact) ([]byte, error) {
	var c bytes.Buffer
	enc := gob.NewEncoder(&c)

	err := enc.Encode(fact)
	if err != nil {
		return nil, err
	}

	return c.Bytes(), nil
}

func decodeFact(item *badger.Item) (*Fact, error) {
	record := Fact{}

	err := item.Value(func(val []byte) error {
		c := bytes.NewBuffer(val)
		dec := gob.NewDecoder(c)
		if err := dec.Decode(&record); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &record, nil
}
