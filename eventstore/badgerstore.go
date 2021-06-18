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
	"runtime"
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
	}
}

func FileStore(path string) EventStore {
	return &BadgerEventStore{
		RootDir:       path,
		MemoryOnly:    false,
		EncryptionKey: nil,
	}
}

func EncryptedFileStore(path string, key []byte, rotationDur time.Duration) EventStore {
	return &BadgerEventStore{
		RootDir:                    path,
		MemoryOnly:                 false,
		EncryptionKey:              key,
		EncryptionRotationDuration: rotationDur,
	}
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

func (b *BadgerEventStore) Register(t interface{}) {
	gob.Register(t)
}

func (b *BadgerEventStore) Append(aggregate string, entity string, content interface{}) (*Tail, error) {
	now := time.Now().UTC()

	tail := Tail{
		Fact: Fact{
			Id:        NewId(now),
			Timestamp: now,
			Content:   content,
		},
	}

	var c bytes.Buffer
	enc := gob.NewEncoder(&c)

	k, err := tail.Fact.Id.MarshalText()
	if err != nil {
		return nil, err
	}

	aggKey := []byte(strings.Join([]string{aggregate, entity}, separator))
	factKey := []byte(strings.Join([]string{aggregate, entity, string(k)}, separator))
	err = enc.Encode(tail.Fact)
	if err != nil {
		return nil, err
	}

	value := c.Bytes()

	db, err := b.kvStore()
	if err != nil {
		return nil, err
	}

	if err = db.Update(func(txn *badger.Txn) error {
		stats := AggregateStats{}
		item, err := txn.Get(aggKey)
		if err == nil {
			// If there is an error then this is a virgin aggregate so there is nothing to read
			err = item.Value(func(val []byte) error {
				dec := gob.NewDecoder(bytes.NewBuffer(val))
				return dec.Decode(&stats)
			})

			if err != nil {
				return err
			}
		}

		err = txn.Set(factKey, value)
		if err != nil {
			return err
		}

		stats.LastId = tail.Fact.Id
		stats.Total += 1
		tail.Total = stats.Total

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err = enc.Encode(&stats)

		if err != nil {
			return err
		}

		return txn.Set(aggKey, buf.Bytes())
	}); err != nil {
		return nil, err
	}

	if runtime.GOOS == "windows" {
		// Windows is not officially supported for badger, but this delay seems to work.
		time.Sleep(1 * time.Millisecond)
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
		stats := AggregateStats{}
		item, err := txn.Get(aggKey)
		if err != nil {
			return err
		}

		// If there is an error then this is a virgin aggregate so there is nothing to read
		err = item.Value(func(val []byte) error {
			dec := gob.NewDecoder(bytes.NewBuffer(val))
			return dec.Decode(&stats)
		})
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

			err := item.Value(func(val []byte) error {
				c := bytes.NewBuffer(val)
				dec := gob.NewDecoder(c)
				var record Fact
				if err = dec.Decode(&record); err != nil {
					return err
				}

				recordId, _ := record.Id.MarshalText()
				if string(recordId) <= factId {
					return nil
				}

				records.List = append(records.List, record)
				return nil
			})

			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &records, nil
}

func (b *BadgerEventStore) Tail(aggregate string, entitty string) (*Tail, error) {
	db, err := b.kvStore()
	if err != nil {
		return nil, err
	}

	tail := Tail{}
	aggKey := strings.Join([]string{aggregate, entitty}, separator)
	err = db.View(func(txn *badger.Txn) error {
		stats := AggregateStats{}
		item, err := txn.Get([]byte(aggKey))
		if err != nil {
			return err
		}

		// If there is an error then this is a virgin aggregate so there is nothing to read
		err = item.Value(func(val []byte) error {
			dec := gob.NewDecoder(bytes.NewBuffer(val))
			return dec.Decode(&stats)
		})
		if err != nil {
			return err
		}

		tail.Total = stats.Total
		k, err := stats.LastId.MarshalText()
		if err != nil {
			return err
		}

		evtKey := []byte(strings.Join([]string{aggregate, entitty, string(k)}, separator))
		item, err = txn.Get(evtKey)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			c := bytes.NewBuffer(val)
			dec := gob.NewDecoder(c)
			return dec.Decode(&tail.Fact)
		})

		return err
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

		time.Sleep(1 * time.Millisecond)
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
