package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/alexandermac/gom"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func main() {
	flags := flag.NewFlagSet("gom", flag.ExitOnError)
	flags.Usage = usage
	help := flags.Bool("help", false, "Show help")
	dir := flags.String("dir", gom.DefaultMigrationsDir, "Migrations directory")
	name := flags.String("name", "", "New migration name")

	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Unable to parse args: %v", err)
		return
	}
	args := flags.Args()

	if *help {
		flags.Usage()
		os.Exit(0)
	}
	if len(args) < 3 {
		flags.Usage()
		os.Exit(1)
	}

	dbDriver := args[0]
	dbString := args[1]
	command := args[2]

	if dbDriver != "sqlite3" {
		log.Fatal("Unsupported dbDriver, gom supports sqlite3 driver only")
	}
	if dbString == "" {
		log.Fatal("dbString must be provided")
	}

	db, err := connectDatabase(dbString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	gom.SetMigrationsDir(*dir)

	switch command {
	case "init":
		if err := gomInit(db, *dir); err != nil {
			log.Fatalf("Error on gom init: %v", err)
		}
		return
	case "create":
		if err := gom.Create(*dir, *name, ""); err != nil {
			log.Fatalf("Error on gom create: %v", err)
		}
		return
	case "migrate":
		if err := gom.Migrate(db); err != nil {
			log.Fatalf("Error on gom migrate: %v", err)
		}
		return
	case "rollback":
		if err := gom.Rollback(db); err != nil {
			log.Fatalf("Error on gom rollback: %v", err)
		}
		return
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func usage() {
	const usagePrefix = `Usage: gom [OPTIONS] DRIVER DBSTRING COMMAND

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

`

	fmt.Print(usagePrefix)
	flag.PrintDefaults()
}

func connectDatabase(dbString string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite", dbString)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
PRAGMA foreign_keys = on;
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func gomInit(db *sqlx.DB, dir string) error {
	const sqlMigrationTemplate = `--
-- This file was automatically created running gom init. You can delete this file if you're familiar with gom.
--
-- A single gom .sql file holds both Up and Down migrations.
-- 
-- All gom .sql files are expected to have a -- migration:up annotation.
-- The -- migration:down annotation is optional, but recommended, and must come after the Up annotation.

-- migration:up
SELECT 'up SQL query';

-- migration:down
SELECT 'down SQL query';
`

	const sqlCreateMigrationsTable = `
CREATE TABLE migrations (
  name text NOT NULL PRIMARY KEY,
  applied_at text NOT NULL
) WITHOUT ROWID;
`

	_, err := os.Stat(dir)
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err == nil, errors.Is(err, fs.ErrExist):
		return fmt.Errorf("The migration directory '%s' already exists", dir)
	default:
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err = gom.Create(dir, "initial", sqlMigrationTemplate); err != nil {
		return err
	}

	if _, err = db.Exec(sqlCreateMigrationsTable); err != nil {
		return err
	}

	log.Println("Gom initialized successfully")
	return nil
}
