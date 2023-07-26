/*
Package Tango allows to attach key-value settings to entities.

An entity is identified by a compound key: universe ID and entity ID. This
allows an entity to coexist at the same time on multiple universes using the
same ID, provided each universe has its own ID too. For instance:

  - Universe ID may be a specific chatroom and entity ID may be the user
    participating in a chatroom. A user may be part of multiple chatrooms
    managed by Tango.
  - Universe ID may be the ID of a specific server or conversation ID, and
    Entity ID may be the ID of the user participating in a conversation,
    allowing the same user to talk on multiple conversations.

Every entity holds a tagbag, which is a dictionary. Multiple tags can be
attached in the same tagbag provided they have different key names, therefore
an entity may have different properties. At the same time, the same dictionary
key may exist for different entities, but because it is part of different
tagbags, each one can have a different value.

# Usage

The package can be obtained with the `go get` command:

	go get gopkg.makigas.es/tango

To use the tag database, you need to provide a database. Note that, however,
most probably this database should be of type SQLite. I haven't tested whether
this package will work with other database engines. The database provided as
a parameter should have the following schema:

	CREATE TABLE IF NOT EXISTS tags(
		id INTEGER PRIMARY KEY,
		universe VARCHAR(64) NOT NULL,
		entity VARCHAR(64) NOT NULL,
		key VARCHAR(64) NOT NULL,
		value TEXT
	);
	CREATE INDEX IF NOT EXISTS tags_entities ON TAGS(universe, entity);
	CREATE UNIQUE INDEX IF NOT EXISTS tags_id ON tags(universe, entity, key);

# Open Source Policy

This package has been made open source in the hope that it is useful for
people studying the behaviour of this software or the programming language or
library set.

However, this is not an open effort. Therefore, issues and pull requests may
be ignored. This program was designed to fulfill some specific requirements
that may not fit the requirements of other people. If other people is reading
this and considering that the application does not behave as expected, they
are free to write their own integrations.
*/
package tango

import (
	"database/sql"
	"encoding/json"
)

// A Tag is a piece of metadata attached to an entity. The Tag interface
// provides methods to extract or modify the value associated with a specific
// tag in the entity dictionary.
type Tag struct {
	db       *sql.DB
	universe string
	entity   string
	key      string
}

var (
	tagUpsert = `
	INSERT INTO tags (universe, entity, key, value) VALUES(?, ?, ?, ?)
	ON CONFLICT(universe, entity, key) DO UPDATE SET value=excluded.value
`
	tagQuery  = `SELECT value FROM tags WHERE universe = ? AND entity = ? AND key = ?`
	tagDelete = `DELETE FROM tags WHERE universe = ? AND entity = ? AND key = ?`

	tagKeys = `SELECT key FROM tags WHERE universe = ? AND entity = ?`
)

// Get the current value of the tag from the persistence. If the tag
// database has a tag for this, it will put the value into the out
// variable and return true. Otherwise, this method returns false.
func (tag *Tag) Get(out any) (bool, error) {
	// Prepare the statement and fetch the results.
	stmt, err := tag.db.Prepare(tagQuery)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	rs, err := stmt.Query(tag.universe, tag.entity, tag.key)
	if err != nil {
		return false, err
	}
	defer rs.Close()

	// if Next() returns true, we have a result. Otherwise, we just haven't.
	if !rs.Next() {
		return false, nil
	}

	// Get the JSON representation of whatever is stored in the database.
	var raw string
	if err := rs.Scan(&raw); err != nil {
		return false, err
	}

	// Convert the raw string into the proper datatype.
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		// IOError
		return false, err
	}
	return true, nil
}

// Set the value of the tag in the persistence engine. After calling
// this method, the value will be persisted into the value of the tag.
// Any other error will be reported.
func (tag *Tag) Set(value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	rawJson := string(raw)
	tx, err := tag.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(tagUpsert)
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(tag.universe, tag.entity, tag.key, rawJson); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// Delete the value of the tag, if such is set. This method should
// fail silently if the persistence lacks the key already.
func (tag *Tag) Delete() error {
	tx, err := tag.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(tagDelete)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(tag.universe, tag.entity, tag.key); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// A TagBag is a collection of tags attached to an entity.
type TagBag struct {
	db       *sql.DB
	universe string
	entity   string
}

// Tag returns a particular tag from the entity given the name of the tag.
func (bag *TagBag) Tag(key string) *Tag {
	return &Tag{db: bag.db, universe: bag.universe, entity: bag.entity, key: key}
}

// Tags returns a list of all the tags in the current tagbag.
func (bag *TagBag) Tags() ([]string, error) {
	stmt, err := bag.db.Prepare(tagKeys)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rs, err := stmt.Query(bag.universe, bag.entity)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	result := []string{}
	for rs.Next() {
		var value string
		rs.Scan(&value)
		result = append(result, value)
	}
	return result, nil
}

type Tags struct {
	db *sql.DB
}

// TagBag returns the proper tagbag collection for a given entity part of an
// universe. Since the actual key for each dictionary is compound of universe
// and entity, calling this method reusing one of the parameters but keeping
// the other one constant, will yield different dictionaries.
func (tags *Tags) TagBag(universe, entity string) *TagBag {
	return &TagBag{db: tags.db, universe: universe, entity: entity}
}

// Tag is a shortcut to get a specific tag for a specific compound key made
// of an entity belonging to a specific universe.
func (tags *Tags) Tag(universe, entity, key string) *Tag {
	return tags.TagBag(universe, entity).Tag(key)
}

// NewTagsEngine returns a valid tags manager that persist into the given
// database. Note that while the function accepts a generic sql.DB object,
// it requires a migration that
func NewTagsEngine(db *sql.DB) *Tags {
	return &Tags{db}
}
