package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"database/sql"

	"github.com/alexandermac/gom"
	_ "modernc.org/sqlite"
)

const VERSION = "0.2.0"

func main() {
	flags := flag.NewFlagSet("gom", flag.ExitOnError)
	flags.Usage = usage
	verbose := flags.Bool("verbose", false, "")
	dir := flags.String("dir", gom.DefaultMigrationsDir, "Migrations directory")
	name := flags.String("name", "", "New migration name")

	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Unable to parse args: %v", err)
		return
	}
	args := flags.Args()

	if len(args) < 1 {
		flags.Usage()
		os.Exit(1)
	}

	firstArg := args[0]
	switch firstArg {
	case "help":
		flags.Usage()
		os.Exit(0)
	case "version":
		fmt.Printf("v%s\n", VERSION)
		os.Exit(0)
	}

	if len(args) < 3 {
		flags.Usage()
		os.Exit(1)
	}

	dbDriver := args[0]
	dbString := args[1]
	command := args[2]

	if dbDriver != "sqlite" {
		log.Fatal("Unsupported dbDriver, gom supports sqlite driver only")
	}
	if dbString == "" {
		log.Fatal("dbString must be provided")
	}

	db, err := connectDatabase(dbDriver, dbString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	gom.SetMigrationsDir(*dir)
	if *verbose {
		gom.SetDefLogLevel(gom.LOG_LEVEL_DEBUG)
	}

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
	log.Println("Done")
}

func usage() {
	const usagePrefix = `Usage: gom [FLAGS] DRIVER DBSTRING COMMAND

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
`

	fmt.Print(usagePrefix)
	flag.PrintDefaults()
}

func connectDatabase(dbDriver string, dbString string) (*sql.DB, error) {
	db, err := sql.Open(dbDriver, dbString)
	if err != nil {
		return nil, err
	}

	if dbDriver == "sqlite" {
		_, err = db.Exec("PRAGMA foreign_keys = on;")
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func gomInit(db *sql.DB, dir string) error {
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
