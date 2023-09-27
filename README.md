# gom

[![Build Status](https://github.com/AlexanderMac/gom/actions/workflows/ci.yml/badge.svg)](https://github.com/AlexanderMac/gom/actions/workflows/ci.yml)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://pkg.go.dev/badge/github.com/alexandermac/gom)](https://pkg.go.dev/github.com/alexandermac/gom)

Gom is a database migration tool, it uses embedding SQL migrations. Requires Go v1.16 or higher.

### Features
TODO
- Golang v1.21

### Install
```sh
# To install the gom binary to your $GOPATH/bin directory
go install github.com/alexandermac/gom/cmd/gom
```

### Usage

##### CLI
```
gom [OPTIONS] DRIVER DBSTRING COMMAND

Commands:
  init                 Creates the migration directory with a sample migration file and the migrations table in the database
  create               Creates a new migration file
  migrate              Migrates the DB to the most recent version available
  rollback             Roll backs the version by 1

Drivers:
  sqlite3

Options:
  --dir                Migrations directory name (absolute or relative path)
  --name               A new migration file suffix

Examples:
  gom -dir db_migrations sqlite3 ./foo.db init
  gom -name create_table sqlite3 ./foo.db create
  gom sqlite3 ./foo.db migrate
  gom sqlite3 ./foo.db rollback
```

##### Embedded migrations

It's possible to embed sql files into binary and corresponding filesystem abstraction. Such migrations can be applied when the app starts.
```go
package main

import (
    "database/sql"
    "embed"

    "github.com/alexandermac/gom"
)

//go:embed my_migrations
var migrationsDir embed.FS

func main() {
	// connect the database

	log.Println("Migrating the database")
	gom.SetBaseFS(migrationsDir)
	gom.SetMigrationsDir("my_migrations")
	if err := gom.Migrate(db); err != nil {
		panic(err)
	}
}
```

### API

##### `func SetBaseFS(fsys simpleFS)`
Sets a base file system to discover migrations. Call this function to pass an embedded migrations variable.

##### `func SetMigrationsDir(dir string)`
Sets the migrations directory.

##### `func SetLogger(l Logger)`
Sets the logger. Must be compatible with gom.Logger interface.

##### `func Create(dir, name, content string) error`
Creates a new migration file. Used in CLI tool.

##### `func Migrate(db *sqlx.DB) error`
Migrates the DB to the most recent version available.

##### `func Rollback(db *sqlx.DB) error`
Roll backs the version by 1.

### License
Licensed under the MIT license.

### Author
Alexander Mac
