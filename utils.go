package gom

import (
	"database/sql"
	"errors"
	"os"
	"path"
	"strings"
)

func createMigrationsTableIfNeeded(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS migrations (
  name text NOT NULL PRIMARY KEY,
  applied_at text NOT NULL
) WITHOUT ROWID;
	`)
	return err
}

func readFileMigrationContent(fileMigration *_FileMigration, mType int) (string, error) {
	mContent, err := baseFS.ReadFile(path.Join(migrationsDir, fileMigration.fileName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}

	mContentByType := getMigrationContentByType(string(mContent), mType)
	return mContentByType, nil
}

func getMigrationContentByType(mContent string, mType int) string {
	hasUp := strings.Contains(mContent, upComment)
	hasDown := strings.Contains(mContent, downComment)
	logger.Debugf("Getting content for type=%d, hasUp=%v, hasDown=%v", mType, hasUp, hasDown)

	if mType == upMigrationType {
		if hasUp {
			if hasDown {
				parts := strings.Split(mContent, downComment)
				logger.Debugf("Getting content for down, parts=%v", parts)
				return parts[0]
			}
			return mContent
		}
	} else {
		if hasDown {
			if hasUp {
				parts := strings.Split(mContent, downComment)
				logger.Debugf("Getting content for up, parts=%v", parts)
				return parts[1]
			}
			return mContent
		}
	}

	return ""
}

func diff[T comparable](ssA, ssB []T, diffing func([]T, T) bool) []T {
	if diffing == nil {
		diffing = func(ss []T, v T) bool {
			index := -1
			for i := range ss {
				if ss[i] == v {
					index = i
					break
				}
			}
			return index == -1
		}
	}

	var ret []T
	for i := range ssA {
		if diffing(ssB, ssA[i]) {
			ret = append(ret, ssA[i])
		}
	}

	return ret
}

func mapColl[T any, R any](ss []T, mapping func(T) R) []R {
	if mapping == nil {
		return make([]R, 0)
	}

	ret := make([]R, len(ss))
	for i := range ss {
		ret[i] = mapping(ss[i])
	}

	return ret
}

func find[T comparable](ss []T, finding func(T) bool) T {
	var ret T
	if finding == nil {
		return ret
	}

	for i := range ss {
		if finding(ss[i]) {
			return ss[i]
		}
	}

	return ret
}
