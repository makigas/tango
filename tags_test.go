package tango

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func prepareTagEngine() (*sql.DB, *Tags, error) {
	// Database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, nil, err
	}

	// Apply migration
	schema := `
	CREATE TABLE IF NOT EXISTS tags(
		id INTEGER PRIMARY KEY,
		universe VARCHAR(64) NOT NULL,
		entity VARCHAR(64) NOT NULL,
		key VARCHAR(64) NOT NULL,
		value TEXT
	);
	CREATE INDEX IF NOT EXISTS tags_entities ON TAGS(universe, entity);
	CREATE UNIQUE INDEX IF NOT EXISTS tags_id ON tags(universe, entity, key);`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, nil, err
	}

	// Create engine and return.
	tags := NewTagsEngine(db)
	return db, tags, nil
}

func TestTagList(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'string', '"hello"')`); err != nil {
		t.Error(err)
	}
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'number', '14')`); err != nil {
		t.Error(err)
	}

	bag := tags.TagBag("1234", "5678")
	list, err := bag.Tags()
	if err != nil {
		t.Error(err)
	}
	expected := []string{"number", "string"}
	if len(expected) != len(list) {
		t.Errorf("Expected list to have length %d, was %d", len(expected), len(list))
	}
	for i, r := range expected {
		if list[i] != r {
			t.Errorf("Expected item %d to be %s, was %s", i, r, list[i])
		}
	}
}

func TestTagsGetNotFound(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Get the tag
	var result *string
	tag := tags.Tag("1234", "5678", "empty")
	exists, err := tag.Get(&result)
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Errorf("Expected key not to exist")
	}
}

func TestTagsGetString(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a string
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'string', '"hello"')`); err != nil {
		t.Error(err)
	}

	// Get the tag
	var result string
	tag := tags.Tag("1234", "5678", "string")
	exists, err := tag.Get(&result)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("Expected key to exist")
	}
	if result != "hello" {
		t.Errorf("Expected key to resolve to 'hello', was `%s`", result)
	}
}

func TestTagsGetNumber(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a string
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'integer', '33')`); err != nil {
		t.Error(err)
	}

	// Get the tag
	var result int
	tag := tags.Tag("1234", "5678", "integer")
	exists, err := tag.Get(&result)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("Expected key to exist")
	}
	if result != 33 {
		t.Errorf("Expected key to resolve to integer 33, was `%d`", result)
	}
}

func TestTagsGetTrue(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a true boolean
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'truthy', 'true')`); err != nil {
		t.Error(err)
	}

	// Get the tag
	var result bool = false
	tag := tags.Tag("1234", "5678", "truthy")
	exists, err := tag.Get(&result)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("Expected key to exist")
	}
	if !result {
		t.Errorf("Expected key to yield true, yielded false")
	}
}

func TestTagsGetFalse(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a false boolean
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'falsy', 'false')`); err != nil {
		t.Error(err)
	}

	// Get the tag
	var result bool = true
	tag := tags.Tag("1234", "5678", "falsy")
	exists, err := tag.Get(&result)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("Expected key to exist")
	}
	if result {
		t.Errorf("Expected key to yield false, yielded true")
	}
}

func TestTagsGetArray(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a false boolean
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'array', '["string", 1234, true]')`); err != nil {
		t.Error(err)
	}

	// Get the tag
	var result []any
	var expected []any = []any{"string", float64(1234), true}
	tag := tags.Tag("1234", "5678", "array")
	exists, err := tag.Get(&result)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("Expected key to exist")
	}
	if len(result) != 3 {
		t.Errorf("Expected key to yield something of length %d, was %d", len(expected), len(result))
	}
	for i, r := range expected {
		if result[i] != r {
			t.Errorf("Expected item %d of array to be %v, was %v", i, r, result[i])
		}
	}
}

func TestTagsGetObject(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a false boolean
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'obj', '{"type": "GuildMember", "id": "12345"}')`); err != nil {
		t.Error(err)
	}

	// Get the tag
	var result map[string]any
	expected := map[string]any{
		"type": "GuildMember",
		"id":   "12345",
	}
	tag := tags.Tag("1234", "5678", "obj")
	exists, err := tag.Get(&result)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("Expected key to exist")
	}
	if len(result) != len(expected) {
		t.Errorf("Expected key to yield something of length %d, was %d", len(expected), len(result))
	}
	for k, v := range expected {
		if result[k] != v {
			t.Errorf("Expected key %s of result to be %v, was %v", k, v, result[k])
		}
	}
}

func TestTagsGetNull(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a false boolean
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'str', 'null')`); err != nil {
		t.Error(err)
	}

	// Get the tag
	fake := "fakeResult"
	var result *string = &fake
	tag := tags.Tag("1234", "5678", "str")
	exists, err := tag.Get(&result)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("Expected key to exist")
	}
	if result != nil {
		t.Errorf("Expected key to yield a null value, yielded %s", *result)
	}
}

func TestTagsInsertString(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set the tag
	tag := tags.Tag("1234", "5678", "string")
	if err := tag.Set("hello"); err != nil {
		t.Error(err)
	}

	// Query the database
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'string'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	var outcome string
	expected := `"hello"`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsInsertNumber(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set the tag
	tag := tags.Tag("1234", "5678", "number")
	if err := tag.Set(33); err != nil {
		t.Error(err)
	}

	// Query the database
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'number'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	var outcome string
	expected := `33`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsInsertTrue(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set the tag
	tag := tags.Tag("1234", "5678", "truthy")
	if err := tag.Set(true); err != nil {
		t.Error(err)
	}

	// Query the database
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'truthy'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	var outcome string
	expected := `true`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsInsertFalse(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set the tag
	tag := tags.Tag("1234", "5678", "falsy")
	if err := tag.Set(false); err != nil {
		t.Error(err)
	}

	// Query the database
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'falsy'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	var outcome string
	expected := `false`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsInsertArray(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set the tag
	tag := tags.Tag("1234", "5678", "array")
	if err := tag.Set([]string{"hello", "world"}); err != nil {
		t.Error(err)
	}

	// Query the database
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'array'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	var outcome string
	expected := `["hello","world"]`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsInsertObject(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set the tag
	tag := tags.Tag("1234", "5678", "object")
	if err := tag.Set(map[string]string{
		"hello": "world",
		"john":  "doe",
	}); err != nil {
		t.Error(err)
	}

	// Query the database
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'object'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	var outcome string
	expected := `{"hello":"world","john":"doe"}`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsInsertNull(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set the tag
	tag := tags.Tag("1234", "5678", "nully")
	if err := tag.Set(nil); err != nil {
		t.Error(err)
	}

	// Query the database
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'nully'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	var outcome string
	expected := `null`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsUpsertString(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a string
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'string', '"hello"')`); err != nil {
		t.Error(err)
	}

	// The string should match 'hello' at the moment.
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'string'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	var outcome string
	expected := `"hello"`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
	result.Close()

	// Then we upsert.
	if err := tags.Tag("1234", "5678", "string").Set("world"); err != nil {
		t.Error(err)
	}

	// Then assert the string changed to world.
	result, err = db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	expected = `"world"`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsUpsertNumber(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a string
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'number', '1234')`); err != nil {
		t.Error(err)
	}

	// The string should match 'hello' at the moment.
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'number'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	var outcome string
	expected := `1234`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
	result.Close()

	// Then we upsert.
	if err := tags.Tag("1234", "5678", "number").Set(5678); err != nil {
		t.Error(err)
	}

	// Then assert the string changed to world.
	result, err = db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	expected = `5678`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsUpsertTrue(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a string
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'bool', 'false')`); err != nil {
		t.Error(err)
	}

	// The string should match 'hello' at the moment.
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'bool'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	var outcome string
	expected := `false`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
	result.Close()

	// Then we upsert.
	if err := tags.Tag("1234", "5678", "bool").Set(true); err != nil {
		t.Error(err)
	}

	// Then assert the string changed to world.
	result, err = db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	expected = `true`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsUpsertFalse(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a string
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'bool', 'true')`); err != nil {
		t.Error(err)
	}

	// The string should match 'hello' at the moment.
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'bool'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	var outcome string
	expected := `true`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
	result.Close()

	// Then we upsert.
	if err := tags.Tag("1234", "5678", "bool").Set(false); err != nil {
		t.Error(err)
	}

	// Then assert the string changed to world.
	result, err = db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	expected = `false`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsUpsertArray(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a string
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'array', '["hello","world"]')`); err != nil {
		t.Error(err)
	}

	// The string should match 'hello' at the moment.
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'array'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	var outcome string
	expected := `["hello","world"]`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
	result.Close()

	// Then we upsert.
	if err := tags.Tag("1234", "5678", "array").Set([]any{"foo", "bar", 123}); err != nil {
		t.Error(err)
	}

	// Then assert the string changed to world.
	result, err = db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	expected = `["foo","bar",123]`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsUpsertObject(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a string
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'objk', '{"hello":"world"}')`); err != nil {
		t.Error(err)
	}

	// The string should match 'hello' at the moment.
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'objk'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	var outcome string
	expected := `{"hello":"world"}`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
	result.Close()

	// Then we upsert.
	if err := tags.Tag("1234", "5678", "objk").Set(map[string]any{"user": "john.doe", "level": 1}); err != nil {
		t.Error(err)
	}

	// Then assert the string changed to world.
	result, err = db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	expected = `{"level":1,"user":"john.doe"}`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsUpsertNullToNotNull(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a string
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'thing', 'null')`); err != nil {
		t.Error(err)
	}

	// The string should match 'hello' at the moment.
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'thing'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	var outcome string
	expected := `null`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
	result.Close()

	// Then we upsert.
	if err := tags.Tag("1234", "5678", "thing").Set("foobar"); err != nil {
		t.Error(err)
	}

	// Then assert the string changed to world.
	result, err = db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	expected = `"foobar"`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsUpsertNotNullToNull(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Insert a string
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'thing', '"foobar"')`); err != nil {
		t.Error(err)
	}

	// The string should match 'hello' at the moment.
	query := `SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'thing'`
	result, err := db.Query(query)
	if err != nil {
		t.Error(err)
	}
	var outcome string
	expected := `"foobar"`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
	result.Close()

	// Then we upsert.
	tags.Tag("1234", "5678", "thing").Set(nil)

	// Then assert the string changed to world.
	result, err = db.Query(query)
	if err != nil {
		t.Error(err)
	}
	defer result.Close()
	expected = `null`
	if !result.Next() {
		t.Errorf("Was not persisted into database")
	}
	result.Scan(&outcome)
	if outcome != expected {
		t.Errorf("Did not persist %s, persisted %s", expected, outcome)
	}
}

func TestTagsDeleteExisting(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Force insert something
	if _, err := db.Exec(`INSERT INTO tags(universe, entity, key, value) VALUES ('1234', '5678', 'bool', 'true')`); err != nil {
		t.Error(err)
	}

	// Should exist in the database
	rs, err := db.Query("SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'bool'")
	if err != nil {
		t.Error(err)
	}
	if !rs.Next() {
		t.Errorf("expected key to already exist")
	}
	rs.Close()

	// Remove it
	if err := tags.Tag("1234", "5678", "bool").Delete(); err != nil {
		t.Error(err)
	}

	// Should not exist anymore
	rs, err = db.Query("SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'bool'")
	if err != nil {
		t.Error(err)
	}
	if rs.Next() {
		t.Errorf("expected key to not already exist")
	}
	rs.Close()
}

func TestTagsDeleteNonExisting(t *testing.T) {
	db, tags, err := prepareTagEngine()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Should not exist in the database
	rs, err := db.Query("SELECT value FROM tags WHERE universe = '1234' AND entity = '5678' AND key = 'bool'")
	if err != nil {
		t.Error(err)
	}
	if rs.Next() {
		t.Errorf("expected key to not already exist")
	}
	rs.Close()

	// Removing it should not fail
	if err := tags.Tag("1234", "5678", "bool").Delete(); err != nil {
		t.Error(err)
	}
}
