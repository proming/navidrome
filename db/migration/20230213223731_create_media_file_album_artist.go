package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up20230213223731, Down20230213223731)
}

func Up20230213223731(tx *sql.Tx) error {
	_, err := tx.Exec(`
	
	  CREATE TABLE media_file_artist_list (
		media_file_id varchar(255),
		artist_id varchar(255),
		artist_name varchar(255),
		FOREIGN KEY (media_file_id) REFERENCES media_file (id) ON DELETE CASCADE,
		CONSTRAINT media_file_artist_ux UNIQUE (media_file_id ASC, artist_id ASC)
	  );
	  
	  CREATE TABLE album_artist_list (
		album_id varchar(255),
		artist_id varchar(255),
		artist_name varchar(255),
		FOREIGN KEY (album_id) REFERENCES album (id) ON DELETE CASCADE,
		CONSTRAINT album_artist_ux UNIQUE (album_id ASC, artist_id ASC)
	  );

	`)
	if err != nil {
		return err
	}
	notice(tx, "A full rescan needs to be performed to import more tags")
	return forceFullRescan(tx)
}

func Down20230213223731(tx *sql.Tx) error {
	return nil
}
