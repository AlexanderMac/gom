<div align="center">
  <h1>gom</h1>
  <p>A database migration tool for Go</p>
  <p>
    <a href="https://github.com/alexandermac/gom/actions/workflows/ci.yml?query=branch%3Amaster"><img src="https://github.com/alexandermac/gom/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
    <a href="https://goreportcard.com/report/github.com/alexandermac/gom"><img src="https://goreportcard.com/badge/github.com/alexandermac/gom" alt="Go Report Card"></a>
    <a href="https://pkg.go.dev/github.com/alexandermac/gom"><img src="https://pkg.go.dev/badge/github.com/alexandermac/gom.svg" alt="Go Docs"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/alexandermac/gom.svg" alt="License"></a>
    <a href="https://img.shields.io/github/v/tag/alexandermac/gom"><img src="https://img.shields.io/github/v/tag/alexandermac/gom" alt="GitHub tag"></a>
  </p>
</div>

Gom is a database migration tool, it uses embedding SQL migrations. Requires Go v1.16 or higher.

# Contents
- [Contents](#contents)
- [Features](#features)
- [Install](#install)
- [Usage](#usage)
- [API](#api)
- [License](#license)

# Features
- Supports SQLite
- CLI
- Embedded migrations
- Plain SQL for writing schema migrations
- Incremental migration version using timestamps
- Run migrations inside a transaction
- Works in Go v1.18+

# Install
```sh
# Install the gom binary in your $GOPATH/bin directory
go install github.com/alexandermac/gom/cmd/gom
```

# Usage

## CLI
```
gom [FLAGS] DRIVER DBSTRING COMMAND

Flags:
  --dir                Migrations directory name (absolute or relative path)
  --name               A new migration file suffix
  --verbose            Prints debug information

Drivers:
  sqlite

Commands:
  help                 Shows this help
  version              Prints app version
  init                 Creates the migration directory with a sample migration file and the migrations table in the database
  create               Creates a new migration file
  migrate              Migrates the DB to the most recent version available
  rollback             Roll backs the version by 1

Examples:
  gom --dir db_migrations sqlite ./foo.db init
  gom --dir db_migrations --name create_users sqlite ./foo.db create
  gom sqlite ./foo.db migrate
  gom sqlite ./foo.db rollback
```

## Embedded migrations

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

# API

### `func SetBaseFS(fsys simpleFS)`
Sets a base file system to discover migrations. Call this function to pass an embedded migrations variable.

### `func SetMigrationsDir(dir string)`
Sets the migrations directory.

### `func SetLogger(l Logger)`
Sets the logger. Must be compatible with gom.Logger interface.

### `func Create(dir, name, content string) error`
Creates a new migration file. Used in CLI tool.

### `func Migrate(db *sql.DB) error`
Migrates the DB to the most recent version available.

### `func Rollback(db *sql.DB) error`
Roll backs the version by 1.

# License
Licensed under the MIT license.

# Author
Alexander Mac
