Package Tango allows to attach key-value settings to entities.

An entity is identified by a compound key: universe ID and entity ID. This
allows an entity to coexist at the same time on multiple universes using the
same ID, provided each universe has its own ID too. For instance:

  - Universe ID may be a specific chatroom and entity ID may be the user
    participating in a chatroom. A user may be part of multiple chatrooms
    managed by Tango.
  - Universe ID may be the ID of a specific server or conversation ID,
    and Entity ID may be the ID of the user participating in a conversation,
    allowing the same user to talk on multiple conversations.

Every entity holds a tagbag, which is a dictionary. Multiple tags can be
attached in the same tagbag provided they have different key names, therefore an
entity may have different properties. At the same time, the same dictionary key
may exist for different entities, but because it is part of different tagbags,
each one can have a different value.

# Usage

The package can be obtained with the `go get` command:

    go get gopkg.makigas.es/tango

To use the tag database, you need to provide a database. Note that, however,
most probably this database should be of type SQLite. I haven't tested whether
this package will work with other database engines. The database provided as a
parameter should have the following schema:

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

This package has been made open source in the hope that it is useful for people
studying the behaviour of this software or the programming language or library
set.

However, this is not an open effort. Therefore, issues and pull requests may be
ignored. This program was designed to fulfill some specific requirements that
may not fit the requirements of other people. If other people is reading this
and considering that the application does not behave as expected, they are free
to write their own integrations.