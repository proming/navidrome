package migrations

import (
	"database/sql"
	"path/filepath"
	"strings"

	"github.com/navidrome/navidrome/consts"
	"github.com/navidrome/navidrome/log"
	"github.com/pressly/goose"
	"golang.org/x/exp/slices"
)

func init() {
	goose.AddMigration(upChangeImageFilesListSeparator, downChangeImageFilesListSeparator)
}

func upChangeImageFilesListSeparator(tx *sql.Tx) error {
	//nolint:gosec
	rows, err := tx.Query(`select id, image_files from album`)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("update album set image_files = ? where id = ?")
	if err != nil {
		return err
	}

	var id, imageFiles sql.NullString
	for rows.Next() {
		err = rows.Scan(&id, &imageFiles)
		if err != nil {
			return err
		}

		var imageFilesStr string
		if imageFiles.Valid {
			imageFilesStr = imageFiles.String
		} else {
			imageFilesStr = ""
		}

		files := upChangeImageFilesListSeparatorDirs(string(imageFilesStr))
		if files == imageFilesStr {
			continue
		}
		_, err = stmt.Exec(files, id.String)
		if err != nil {
			log.Error("Error updating album's image file list", "files", files, "id", id, err)
		}
	}
	return rows.Err()
}

func upChangeImageFilesListSeparatorDirs(filePaths string) string {
	allPaths := filepath.SplitList(filePaths)
	slices.Sort(allPaths)
	allPaths = slices.Compact(allPaths)
	return strings.Join(allPaths, consts.Zwsp)
}

func downChangeImageFilesListSeparator(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
