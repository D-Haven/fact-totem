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
	separator = "|"
)

type BadgerEventStore struct {
	RootDir                    string
	MemoryOnly                 bool
	EncryptionKey              []byte
	EncryptionRotationDuration time.Duration
	db                         *badger.DB
}

type Record struct {
	Id        ulid.ULID
	Timestamp time.Time
	Content   interface{}
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

	b.Register(Record{})
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	b.db = db
	return b.db, nil
}

func (b *BadgerEventStore) Append(aggregate string, key string, content interface{}) (string, error) {
	now := time.Now().UTC()

	record := Record{
		Id:        NewId(now),
		Timestamp: now,
		Content:   content,
	}

	var c bytes.Buffer
	enc := gob.NewEncoder(&c)

	k, err := record.Id.MarshalText()
	if err != nil {
		return "", err
	}

	evtKey := []byte(strings.Join([]string{aggregate, key, string(k)}, separator))
	err = enc.Encode(record)
	if err != nil {
		return "", err
	}

	value := c.Bytes()

	db, err := b.kvStore()
	if err != nil {
		return "", err
	}

	if err = db.Update(func(txn *badger.Txn) error {
		return txn.Set(evtKey, value)
	}); err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		// Windows is not officially supported for badger, but this delay seems to work.
		time.Sleep(1 * time.Millisecond)
	}

	return string(k), nil
}

func (b *BadgerEventStore) Register(t interface{}) {
	gob.Register(t)
}

func (b *BadgerEventStore) Read(aggregate string, key string) ([]interface{}, string, error) {
	return b.ReadFrom(aggregate, key, "")
}

func (b *BadgerEventStore) ReadFrom(aggregate string, key string, evtId string) ([]interface{}, string, error) {
	db, err := b.kvStore()

	if err != nil {
		return nil, "", err
	}

	prefix := []byte(strings.Join([]string{aggregate, key}, separator))
	var values []interface{}
	var lastEvt string

	if err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		// Walk all the events using the aggregate as a prefix
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				c := bytes.NewBuffer(val)
				dec := gob.NewDecoder(c)
				var record Record
				if err = dec.Decode(&record); err != nil {
					return err
				}

				recordId, _ := record.Id.MarshalText()
				if string(recordId) <= evtId {
					return nil
				}

				lastEvt = string(recordId)
				values = append(values, record.Content)
				return nil
			})

			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, "", err
	}

	return values, lastEvt, nil
}

func (b *BadgerEventStore) Tail(aggregate string, key string) (string, error) {
	db, err := b.kvStore()
	if err != nil {
		return "", err
	}

	var lastKey string
	prefix := strings.Join([]string{aggregate, key}, separator)
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // keys only
		opts.Reverse = true

		it := txn.NewIterator(opts)
		defer it.Close()

		p := []byte(prefix)

		it.Seek(append(p, 0xff))
		if it.ValidForPrefix(p) {
			key := string(it.Item().Key())
			split := strings.Split(key, separator)
			lastKey = split[len(split)-1]
		}

		return nil
	})

	return lastKey, err
}

func (b *BadgerEventStore) ListKeysForAggregate(aggregate string) ([]string, error) {
	prefix := []byte(aggregate)
	var keys []string
	db, err := b.kvStore()

	if err != nil {
		return nil, err
	}

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
				keys = append(keys, key)
				lastKey = key
			}
		}

		if len(lastKey) > 0 && (len(keys) == 0 || keys[len(keys)-1] != lastKey) {
			keys = append(keys, lastKey)
		}

		time.Sleep(1 * time.Millisecond)
		return nil
	}); err != nil {
		return nil, err
	}

	return keys, nil
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
