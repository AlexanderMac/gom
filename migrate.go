package gom

import (
	"database/sql"
	"sort"
	"strings"
)

func Migrate(db *sql.DB) error {
	logger.Info("Migrate")

	if err := createMigrationsTableIfNeeded(db); err != nil {
		return err
	}

	fileUpMigrations, err := getUpFileMigrations()
	if err != nil {
		return err
	}

	dbMigrations, err := getDbMigrations(db)
	if err != nil {
		return err
	}

	fileMigrationNames := mapColl(fileUpMigrations, func(fileMigration _FileMigration) string {
		return fileMigration.name
	})
	dbMigrationNames := mapColl(dbMigrations, func(dbMigration _DbMigration) string {
		return dbMigration.Name
	})
	migrationsDiff := diff(fileMigrationNames, dbMigrationNames, nil)
	if len(migrationsDiff) == 0 {
		logger.Info("No pending migrations")
		return nil
	}
	logger.Infof("Pending migrations: %v", migrationsDiff)

	pendingMigrations := make([]_FileMigration, len(migrationsDiff))
	for i := range migrationsDiff {
		pendingMigrations[i] = find(fileUpMigrations, func(fileMigration _FileMigration) bool {
			return fileMigration.name == migrationsDiff[i]
		})
		fileContent, err := readFileMigrationContent(&pendingMigrations[i], upMigrationType)
		if err != nil {
			return err
		}
		pendingMigrations[i].fileContent = fileContent
	}

	return runMigrations(db, pendingMigrations)
}

func getUpFileMigrations() ([]_FileMigration, error) {
	migrationFiles, err := baseFS.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var fileMigrations []_FileMigration
	for i := range migrationFiles {
		fileName := migrationFiles[i].Name()
		fileNameParts := strings.Split(fileName, ".")
		fileMigration := _FileMigration{
			name:     fileNameParts[0],
			fileName: fileName,
		}
		fileMigrations = append(fileMigrations, fileMigration)
	}

	sort.Slice(fileMigrations, func(i, j int) bool {
		return fileMigrations[i].name < fileMigrations[j].name
	})

	return fileMigrations, nil
}

func getDbMigrations(db *sql.DB) ([]_DbMigration, error) {
	var dbMigrations []_DbMigration
	rows, err := db.Query(`
SELECT name
FROM migrations
ORDER BY 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m _DbMigration
		err = rows.Scan(&m.Name)
		if err != nil {
			return nil, err
		}
		dbMigrations = append(dbMigrations, m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return dbMigrations, nil
}

func runMigrations(db *sql.DB, pendingMigrations []_FileMigration) error {
	for i := range pendingMigrations {
		runningMigration := pendingMigrations[i]
		if err := runSingleMigration(db, runningMigration); err != nil {
			return err
		}
	}

	return nil
}

func runSingleMigration(db *sql.DB, runningMigration _FileMigration) error {
	logger.Infof("Running migration: %s...", runningMigration.name)
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:golint,errcheck

	if _, err = tx.Exec(runningMigration.fileContent); err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO migrations (name, applied_at)
VALUES (?, datetime())
	`, runningMigration.name)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}
