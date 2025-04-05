package gom

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const newMigrationContent = "-- migration:up\n\n-- migration:down"

func Create(dir, name, content string) error {
	version := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s.sql", version, name)
	filePath := filepath.Join(dir, fileName)

	if content == "" {
		content = newMigrationContent
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return err
	}

	logger.Infof("Created a new migration file: %s", filePath)
	return nil
}
