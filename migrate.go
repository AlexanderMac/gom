package gom

import (
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

func Migrate(db *sqlx.DB) error {
	logger.Info("Migrate")

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

	err = runMigrations(db, pendingMigrations)
	return err
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

func getDbMigrations(db *sqlx.DB) ([]_DbMigration, error) {
	var dbMigrations []_DbMigration
	err := db.Select(&dbMigrations, `
SELECT name
FROM migrations
ORDER BY 1
	`)
	if err != nil {
		return nil, err
	}

	return dbMigrations, nil
}

func runMigrations(db *sqlx.DB, pendingMigrations []_FileMigration) error {
	for i := range pendingMigrations {
		runningMigration := pendingMigrations[i]
		logger.Infof("Running migration: %s...", runningMigration.name)
		err := runSingleMigration(db, runningMigration)
		if err != nil {
			return err
		}
	}

	return nil
}

func runSingleMigration(db *sqlx.DB, pendingMigration _FileMigration) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:golint,errcheck

	_, err = tx.Exec(pendingMigration.fileContent)
	if err != nil {
		return err
	}

	_, err = tx.NamedExec(`
INSERT INTO migrations (name, applied_at)
VALUES (:name, datetime())
	`, _DbMigration{Name: pendingMigration.name})
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}
