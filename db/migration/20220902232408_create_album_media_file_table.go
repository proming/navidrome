package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(up20220902232408, down20220902232408)
}

func up20220902232408(tx *sql.Tx) error {
	_, err := tx.Exec(`
    CREATE TABLE album_media_file (
        album_id varchar,
        media_file_id varchar,
        track_number INTEGER,
        disc_number INTEGER,
        FOREIGN KEY (album_id) REFERENCES album (id) ON DELETE CASCADE,
        FOREIGN KEY (media_file_id) REFERENCES media_file (id) ON DELETE CASCADE
      );
      
      CREATE INDEX album_media_file_album_id
      ON album_media_file (
        album_id ASC
      );
      
      CREATE INDEX album_media_file_file_id
      ON album_media_file (
        media_file_id ASC
      );
`)

	return err
}

func down20220902232408(tx *sql.Tx) error {
	return nil
}
