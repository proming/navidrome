package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up20220826082540, Down20220826082540)
}

func Up20220826082540(tx *sql.Tx) error {
	_, err := tx.Exec(`
alter table media_file
	add visible varchar default 'T';
create index if not exists media_file_visible
	on media_file(visible);

alter table media_file
	add all_artist_ids varchar ;
create index if not exists media_file_all_artist_ids
	on media_file(all_artist_ids);

update media_file set visible = 'T', all_artist_ids = artist_id || '/' || album_artist_id;

`)

	return err
}

func Down20220826082540(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
