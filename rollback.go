package gom

import (
	"github.com/jmoiron/sqlx"
)

func Rollback(db *sqlx.DB) error {
	logger.Info("Rollback")

	lastDbMigration, err := getLastDbMigration(db)
	if err != nil {
		return err
	}
	if lastDbMigration.Name == "" {
		logger.Info("No migrations to rollback")
		return nil
	}

	rollingBackMigration := _FileMigration{
		name:     lastDbMigration.Name,
		fileName: lastDbMigration.Name + ".sql",
	}
	fileContent, err := readFileMigrationContent(&rollingBackMigration, downMigrationType)
	if err != nil {
		return err
	}
	rollingBackMigration.fileContent = fileContent
	logger.Infof("Rolling back migration: %s...", rollingBackMigration.name)

	err = rollbackMigration(db, rollingBackMigration)
	return err
}

func getLastDbMigration(db *sqlx.DB) (_DbMigration, error) {
	var dbMigrations []_DbMigration
	err := db.Select(&dbMigrations, `
SELECT name
FROM migrations
ORDER BY 1 DESC
LIMIT 1
	`)
	if err != nil {
		return _DbMigration{}, err
	}

	if len(dbMigrations) > 0 {
		return dbMigrations[0], nil
	}
	return _DbMigration{}, nil
}

func rollbackMigration(db *sqlx.DB, rollingBackMigration _FileMigration) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:golint,errcheck

	if rollingBackMigration.fileContent != "" {
		_, err = tx.Exec(rollingBackMigration.fileContent)
		if err != nil {
			return err
		}
	}

	_, err = tx.NamedExec(`
DELETE FROM migrations
WHERE name = :name
	`, _DbMigration{Name: rollingBackMigration.name})
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}
