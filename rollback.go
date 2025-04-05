package gom

import "database/sql"

func Rollback(db *sql.DB) error {
	logger.Info("Rollback")

	if err := createMigrationsTableIfNeeded(db); err != nil {
		return err
	}

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

	return rollbackMigration(db, rollingBackMigration)
}

func getLastDbMigration(db *sql.DB) (_DbMigration, error) {
	var dbLastMigration _DbMigration
	rows, err := db.Query(`
SELECT name
FROM migrations
ORDER BY 1 DESC
LIMIT 1
	`)
	if err != nil {
		return dbLastMigration, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&dbLastMigration.Name)
		if err != nil {
			return dbLastMigration, err
		}
		break //nolint:golint,staticcheck
	}
	if err = rows.Err(); err != nil {
		return dbLastMigration, err
	}

	return dbLastMigration, nil
}

func rollbackMigration(db *sql.DB, rollingBackMigration _FileMigration) error {
	logger.Infof("Rolling back migration: %s...", rollingBackMigration.name)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:golint,errcheck

	if rollingBackMigration.fileContent != "" {
		if _, err = tx.Exec(rollingBackMigration.fileContent); err != nil {
			return err
		}
	}

	_, err = tx.Exec(`
DELETE FROM migrations
WHERE name = ?
	`, rollingBackMigration.name)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}
