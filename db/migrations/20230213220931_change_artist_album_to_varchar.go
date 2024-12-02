package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up20230213220931, Down20230213220931)
}

func Up20230213220931(_ context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
alter table media_file
    add all_artist_ids varchar;

create index if not exists media_file_all_artist_ids
	on media_file (all_artist_ids);
`)
	if err != nil {
		return err
	}
	notice(tx, "A full rescan needs to be performed to import more tags")
	return forceFullRescan(tx)
}

func Down20230213220931(_ context.Context, tx *sql.Tx) error {
	return nil
}
