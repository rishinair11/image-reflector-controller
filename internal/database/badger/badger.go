/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package badger

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v3"

	"github.com/fluxcd/image-reflector-controller/internal/database"
)

const tagsPrefix = "tags"

// BadgerDatabase provides implementations of the tags database based on Badger.
type BadgerDatabase struct {
	db *badger.DB
}

var _ database.DatabaseWriter = &BadgerDatabase{}
var _ database.DatabaseReader = &BadgerDatabase{}

// NewBadgerDatabase creates and returns a new database implementation using
// Badger for storing the image tags.
func NewBadgerDatabase(db *badger.DB) *BadgerDatabase {
	return &BadgerDatabase{
		db: db,
	}
}

// Tags implements the DatabaseReader interface, fetching the tags for the repo.
//
// If the repo does not exist, an empty set of tags is returned.
func (a *BadgerDatabase) Tags(repo string) ([]database.Tag, error) {
	var tags []database.Tag
	err := a.db.View(func(txn *badger.Txn) error {
		var err error
		tags, err = getOrEmpty(txn, repo)
		return err
	})
	return tags, err
}

// SetTags implements the DatabaseWriter interface, recording the tags against
// the repo.
//
// It overwrites existing tag sets for the provided repo.
func (a *BadgerDatabase) SetTags(repo string, tags []database.Tag) error {
	b, err := marshal(tags)
	if err != nil {
		return err
	}
	return a.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(keyForRepo(tagsPrefix, repo), b)
		return txn.SetEntry(e)
	})
}

func keyForRepo(prefix, repo string) []byte {
	return []byte(fmt.Sprintf("%s:%s", prefix, repo))
}

func getOrEmpty(txn *badger.Txn, repo string) ([]database.Tag, error) {
	item, err := txn.Get(keyForRepo(tagsPrefix, repo))
	if err == badger.ErrKeyNotFound {
		return []database.Tag{}, nil
	}
	if err != nil {
		return nil, err
	}
	var tags []database.Tag
	err = item.Value(func(val []byte) error {
		tags, err = unmarshal(val)
		return err
	})
	return tags, err
}

func marshal(t []database.Tag) ([]byte, error) {
	return json.Marshal(t)
}

func unmarshal(b []byte) ([]database.Tag, error) {
	var tags []database.Tag
	if err := json.Unmarshal(b, &tags); err != nil {
		// If unmarshalling fails we may be operating on an old database so try to read the old format before eventually bailing out.
		var tagsOld []string
		if err2 := json.Unmarshal(b, &tagsOld); err2 != nil {
			return nil, fmt.Errorf("failed unmarshaling values. First error: %s. Second error: %w", err, err2)
		}
		tags = make([]database.Tag, len(tagsOld))
		for idx, tag := range tagsOld {
			tags[idx] = database.Tag{Name: tag}
		}
	}
	return tags, nil
}
