package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up20230213220931, Down20230213220931)
}

func Up20230213220931(tx *sql.Tx) error {
	_, err := tx.Exec(`
	
	  CREATE TABLE media_file_tmp (
		id varchar(255) NOT NULL,
		path varchar(255) NOT NULL DEFAULT '',
		title varchar(255) NOT NULL DEFAULT '',
		album varchar(255) NOT NULL DEFAULT '',
		artist varchar NOT NULL DEFAULT '',
		artist_id varchar NOT NULL DEFAULT '',
		album_artist varchar NOT NULL DEFAULT '',
		album_id varchar(255) NOT NULL DEFAULT '',
		has_cover_art bool NOT NULL DEFAULT FALSE,
		track_number integer NOT NULL DEFAULT 0,
		disc_number integer NOT NULL DEFAULT 0,
		year integer NOT NULL DEFAULT 0,
		size integer NOT NULL DEFAULT 0,
		suffix varchar(255) NOT NULL DEFAULT '',
		duration real NOT NULL DEFAULT 0,
		bit_rate integer NOT NULL DEFAULT 0,
		genre varchar(255) NOT NULL DEFAULT '',
		compilation bool NOT NULL DEFAULT FALSE,
		created_at datetime,
		updated_at datetime,
		full_text varchar DEFAULT '',
		album_artist_id varchar DEFAULT '',
		order_album_name varchar COLLATE nocase,
		order_album_artist_name varchar COLLATE nocase,
		order_artist_name varchar COLLATE nocase,
		sort_album_name varchar COLLATE nocase,
		sort_artist_name varchar COLLATE nocase,
		sort_album_artist_name varchar COLLATE nocase,
		sort_title varchar(255) COLLATE nocase,
		disc_subtitle varchar(255),
		mbz_track_id varchar(255),
		mbz_album_id varchar(255),
		mbz_artist_id varchar(255),
		mbz_album_artist_id varchar(255),
		mbz_album_type varchar(255),
		mbz_album_comment varchar(255),
		catalog_num varchar(255),
		comment varchar,
		lyrics varchar,
		bpm integer,
		channels integer,
		order_title varchar COLLATE NOCASE,
		mbz_release_track_id varchar(255),
		rg_album_gain real,
		rg_album_peak real,
		rg_track_gain real,
		rg_track_peak real,
		all_artist_ids varchar,
		PRIMARY KEY (id)
	  );
	  
	  INSERT INTO media_file_tmp (id, path, title, album, artist, artist_id, album_artist, album_id, has_cover_art, track_number, disc_number, year, size, suffix, duration, bit_rate, genre, compilation, created_at, updated_at, full_text, album_artist_id, order_album_name, order_album_artist_name, order_artist_name, sort_album_name, sort_artist_name, sort_album_artist_name, sort_title, disc_subtitle, mbz_track_id, mbz_album_id, mbz_artist_id, mbz_album_artist_id, mbz_album_type, mbz_album_comment, catalog_num, comment, lyrics, bpm, channels, order_title, mbz_release_track_id, rg_album_gain, rg_album_peak, rg_track_gain, rg_track_peak) SELECT id, path, title, album, artist, artist_id, album_artist, album_id, has_cover_art, track_number, disc_number, year, size, suffix, duration, bit_rate, genre, compilation, created_at, updated_at, full_text, album_artist_id, order_album_name, order_album_artist_name, order_artist_name, sort_album_name, sort_artist_name, sort_album_artist_name, sort_title, disc_subtitle, mbz_track_id, mbz_album_id, mbz_artist_id, mbz_album_artist_id, mbz_album_type, mbz_album_comment, catalog_num, comment, lyrics, bpm, channels, order_title, mbz_release_track_id, rg_album_gain, rg_album_peak, rg_track_gain, rg_track_peak FROM media_file;
	  
	  DROP TABLE media_file;
	  ALTER TABLE media_file_tmp RENAME TO media_file;
		  
	  CREATE INDEX media_file_album_artist
	  ON media_file (
		album_artist ASC
	  );
	  
	  CREATE INDEX media_file_album_id
	  ON media_file (
		album_id ASC
	  );
	  
	  CREATE INDEX media_file_artist
	  ON media_file (
		artist ASC
	  );
	  
	  CREATE INDEX media_file_artist_album_id
	  ON media_file (
		album_artist_id ASC
	  );
	  
	  CREATE INDEX media_file_artist_id
	  ON media_file (
		artist_id ASC
	  );
	  
	  CREATE INDEX media_file_bpm
	  ON media_file (
		bpm ASC
	  );
	  
	  CREATE INDEX media_file_channels
	  ON media_file (
		channels ASC
	  );
	  
	  CREATE INDEX media_file_created_at
	  ON media_file (
		created_at ASC
	  );
	  
	  CREATE INDEX media_file_duration
	  ON media_file (
		duration ASC
	  );
	  
	  CREATE INDEX media_file_full_text
	  ON media_file (
		full_text ASC
	  );
	  
	  CREATE INDEX media_file_genre
	  ON media_file (
		genre ASC
	  );
	  
	  CREATE INDEX media_file_mbz_track_id
	  ON media_file (
		mbz_track_id ASC
	  );
	  
	  CREATE INDEX media_file_order_album_name
	  ON media_file (
		order_album_name ASC
	  );
	  
	  CREATE INDEX media_file_order_artist_name
	  ON media_file (
		order_artist_name ASC
	  );
	  
	  CREATE INDEX media_file_order_title
	  ON media_file (
		order_title ASC
	  );
	  
	  CREATE INDEX media_file_path
	  ON media_file (
		path ASC
	  );
	  
	  CREATE INDEX media_file_title
	  ON media_file (
		title ASC
	  );
	  
	  CREATE INDEX media_file_track_number
	  ON media_file (
		disc_number ASC,
		track_number ASC
	  );
	  
	  CREATE INDEX media_file_updated_at
	  ON media_file (
		updated_at ASC
	  );
	  
	  CREATE INDEX media_file_year
	  ON media_file (
		year ASC
	  );
	  
	  CREATE INDEX media_file_all_artist_ids
	  ON media_file (
		all_artist_ids ASC
	  );
	  
	  
	  CREATE TABLE album_tmp (
		id varchar(255) NOT NULL,
		name varchar(255) NOT NULL DEFAULT '',
		artist_id varchar(255) NOT NULL DEFAULT '',
		embed_art_path varchar(255) NOT NULL DEFAULT '',
		artist varchar NOT NULL DEFAULT '',
		album_artist varchar NOT NULL DEFAULT '',
		min_year int NOT NULL DEFAULT 0,
		max_year integer NOT NULL DEFAULT 0,
		compilation bool NOT NULL DEFAULT FALSE,
		song_count integer NOT NULL DEFAULT 0,
		duration real NOT NULL DEFAULT 0,
		genre varchar(255) NOT NULL DEFAULT '',
		created_at datetime,
		updated_at datetime,
		full_text varchar DEFAULT '',
		album_artist_id varchar DEFAULT '',
		order_album_name varchar COLLATE nocase,
		order_album_artist_name varchar COLLATE nocase,
		sort_album_name varchar COLLATE nocase,
		sort_artist_name varchar COLLATE nocase,
		sort_album_artist_name varchar COLLATE nocase,
		size integer NOT NULL DEFAULT 0,
		mbz_album_id varchar(255),
		mbz_album_artist_id varchar(255),
		mbz_album_type varchar(255),
		mbz_album_comment varchar(255),
		catalog_num varchar(255),
		comment varchar,
		all_artist_ids varchar,
		image_files varchar,
		paths varchar,
		description varchar(255) NOT NULL DEFAULT '',
		small_image_url varchar(255) NOT NULL DEFAULT '',
		medium_image_url varchar(255) NOT NULL DEFAULT '',
		large_image_url varchar(255) NOT NULL DEFAULT '',
		external_url varchar(255) NOT NULL DEFAULT '',
		external_info_updated_at datetime,
		PRIMARY KEY (id)
	  );
	  
	  INSERT INTO album_tmp (id, name, artist_id, embed_art_path, artist, album_artist, min_year, max_year, compilation, song_count, duration, genre, created_at, updated_at, full_text, album_artist_id, order_album_name, order_album_artist_name, sort_album_name, sort_artist_name, sort_album_artist_name, size, mbz_album_id, mbz_album_artist_id, mbz_album_type, mbz_album_comment, catalog_num, comment, all_artist_ids, image_files, paths, description, small_image_url, medium_image_url, large_image_url, external_url, external_info_updated_at) SELECT id, name, artist_id, embed_art_path, artist, album_artist, min_year, max_year, compilation, song_count, duration, genre, created_at, updated_at, full_text, album_artist_id, order_album_name, order_album_artist_name, sort_album_name, sort_artist_name, sort_album_artist_name, size, mbz_album_id, mbz_album_artist_id, mbz_album_type, mbz_album_comment, catalog_num, comment, all_artist_ids, image_files, paths, description, small_image_url, medium_image_url, large_image_url, external_url, external_info_updated_at FROM album;
	  
	  DROP TABLE album;
	  ALTER TABLE album_tmp RENAME TO album;
	  
	  CREATE INDEX album_all_artist_ids
	  ON album (
		all_artist_ids ASC
	  );
	  
	  CREATE INDEX album_alphabetical_by_artist
	  ON album (
		compilation ASC,
		order_album_artist_name ASC,
		order_album_name ASC
	  );
	  
	  CREATE INDEX album_artist
	  ON album (
		artist ASC
	  );
	  
	  CREATE INDEX album_artist_album
	  ON album (
		artist ASC
	  );
	  
	  CREATE INDEX album_artist_album_id
	  ON album (
		album_artist_id ASC
	  );
	  
	  CREATE INDEX album_artist_id
	  ON album (
		artist_id ASC
	  );
	  
	  CREATE INDEX album_created_at
	  ON album (
		created_at ASC
	  );
	  
	  CREATE INDEX album_full_text
	  ON album (
		full_text ASC
	  );
	  
	  CREATE INDEX album_genre
	  ON album (
		genre ASC
	  );
	  
	  CREATE INDEX album_max_year
	  ON album (
		max_year ASC
	  );
	  
	  CREATE INDEX album_mbz_album_type
	  ON album (
		mbz_album_type ASC
	  );
	  
	  CREATE INDEX album_min_year
	  ON album (
		min_year ASC
	  );
	  
	  CREATE INDEX album_name
	  ON album (
		name ASC
	  );
	  
	  CREATE INDEX album_order_album_artist_name
	  ON album (
		order_album_artist_name ASC
	  );
	  
	  CREATE INDEX album_order_album_name
	  ON album (
		order_album_name ASC
	  );
	  
	  CREATE INDEX album_size
	  ON album (
		size ASC
	  );
	  
	  CREATE INDEX album_updated_at
	  ON album (
		updated_at ASC
	  );

`)
	if err != nil {
		return err
	}
	notice(tx, "A full rescan needs to be performed to import more tags")
	return forceFullRescan(tx)
}

func Down20230213220931(tx *sql.Tx) error {
	return nil
}
