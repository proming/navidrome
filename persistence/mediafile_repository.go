package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	. "github.com/Masterminds/squirrel"
	"github.com/beego/beego/v2/client/orm"
	"github.com/deluan/rest"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/utils"
)

type mediaFileRepository struct {
	sqlRepository
	sqlRestful
}

func NewMediaFileRepository(ctx context.Context, o orm.QueryExecutor) *mediaFileRepository {
	r := &mediaFileRepository{}
	r.ctx = ctx
	r.ormer = o
	r.tableName = "media_file"
	r.sortMappings = map[string]string{
		"artist": "order_artist_name asc, order_album_name asc, disc_number asc, track_number asc",
		"album":  "order_album_name asc, disc_number asc, track_number asc, order_artist_name asc, title asc",
		"random": "RANDOM()",
	}
	r.filterMappings = map[string]filterFunc{
		"id":              idFilter(r.tableName),
		"title":           fullTextFilter,
		"starred":         booleanFilter,
		"album_artist_id": albumArtistFilter,
	}
	return r
}

func albumArtistFilter(field string, value interface{}) Sqlizer {
	return Like{"all_artist_ids": fmt.Sprintf("%%%s%%", value)}
}

func (r *mediaFileRepository) CountAll(options ...model.QueryOptions) (int64, error) {
	var sql SelectBuilder
	if len(options) > 0 && options[0].Filters != nil {
		s, _, _ := options[0].Filters.ToSql()
		if strings.Contains(s, "album_id") {
			sql = r.newSelectWithAnnotationContainAlbum("media_file.id").Where(Eq{"visible": "T"})
		}
	}

	if sql == (SelectBuilder{}) {
		sql = r.newSelectWithAnnotation("media_file.id").Where(Eq{"visible": "T"})
		// sql := r.newSelectWithAnnotation("media_file.id").Where(Eq{"visible": "T"})
	}
	sql = r.withGenres(sql)
	return r.count(sql, options...)
}

func (r *mediaFileRepository) Exists(id string) (bool, error) {
	return r.exists(Select().Where(Eq{"media_file.id": id}))
}

func (r *mediaFileRepository) Put(m *model.MediaFile) error {
	m.FullText = getFullText(m.Title, m.Album, utils.SplitAndJoinStrings(m.Artist), utils.SplitAndJoinStrings(m.AlbumArtist),
		m.SortTitle, m.SortAlbumName, utils.SplitAndJoinStrings(m.SortArtistName), utils.SplitAndJoinStrings(m.SortAlbumArtistName), m.DiscSubtitle)
	m.AllArtistIDs = utils.SanitizeStrings(strings.ReplaceAll(m.ArtistID, "/", " "), strings.ReplaceAll(m.AlbumArtistID, "/", " "))
	if len(m.Visible) < 1 {
		m.Visible = "T"
	}
	_, err := r.put(m.ID, m)
	if err != nil {
		return err
	}
	return r.updateGenres(m.ID, r.tableName, m.Genres)
}

func (r *mediaFileRepository) selectMediaFile(options ...model.QueryOptions) SelectBuilder {
	var sql SelectBuilder
	if len(options) > 0 && options[0].Filters != nil {
		s, _, _ := options[0].Filters.ToSql()
		if strings.Contains(s, "album_id") {
			sql = r.newSelectWithAnnotationContainAlbum("media_file.id", options...).Columns("media_file.*")
		}
	}

	if sql == (SelectBuilder{}) {
		sql = r.newSelectWithAnnotation("media_file.id", options...).Columns("media_file.*")
	}

	sql = r.withBookmark(sql, "media_file.id")
	if len(options) > 0 && options[0].Filters != nil {
		s, _, _ := options[0].Filters.ToSql()
		// If there's any reference of genre in the filter, joins with genre
		if strings.Contains(s, "genre") {
			sql = r.withGenres(sql)
			// If there's no filter on genre_id, group the results by media_file.id
			if !strings.Contains(s, "genre_id") {
				sql = sql.GroupBy("media_file.id")
			}
		}
	}
	return sql
}

func (r *mediaFileRepository) Get(id string) (*model.MediaFile, error) {
	sel := r.selectMediaFile().Where(Eq{"media_file.id": id})
	var res model.MediaFiles
	if err := r.queryAll(sel, &res); err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, model.ErrNotFound
	}
	err := r.loadMediaFileGenres(&res)
	return &res[0], err
}

func (r *mediaFileRepository) GetAll(options ...model.QueryOptions) (model.MediaFiles, error) {
	sq := r.selectMediaFile(options...).Where(Eq{"visible": "T"})
	res := model.MediaFiles{}
	err := r.queryAll(sq, &res)
	if err != nil {
		return nil, err
	}
	err = r.loadMediaFileGenres(&res)
	return res, err
}

func (r *mediaFileRepository) FindByPath(path string) (*model.MediaFile, error) {
	sel := r.newSelect().Columns("*").Where(Eq{"path": path})
	var res model.MediaFiles
	if err := r.queryAll(sel, &res); err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, model.ErrNotFound
	}
	return &res[0], nil
}

func cleanPath(path string) string {
	path = filepath.Clean(path)
	if !strings.HasSuffix(path, string(os.PathSeparator)) {
		path += string(os.PathSeparator)
	}
	return path
}

func pathStartsWith(path string) Eq {
	substr := fmt.Sprintf("substr(path, 1, %d)", utf8.RuneCountInString(path))
	return Eq{substr: path}
}

// FindAllByPath only return mediafiles that are direct children of requested path
func (r *mediaFileRepository) FindAllByPath(path string) (model.MediaFiles, error) {
	// Query by path based on https://stackoverflow.com/a/13911906/653632
	path = cleanPath(path)
	pathLen := utf8.RuneCountInString(path)
	sel0 := r.newSelect().Columns("media_file.*", fmt.Sprintf("substr(path, %d) AS item", pathLen+2)).
		Where(pathStartsWith(path))
	sel := r.newSelect().Columns("*", "item NOT GLOB '*"+string(os.PathSeparator)+"*' AS isLast").
		Where(Eq{"isLast": 1}).FromSelect(sel0, "sel0")

	res := model.MediaFiles{}
	err := r.queryAll(sel, &res)
	return res, err
}

// FindPathsRecursively returns a list of all subfolders of basePath, recursively
func (r *mediaFileRepository) FindPathsRecursively(basePath string) ([]string, error) {
	path := cleanPath(basePath)
	// Query based on https://stackoverflow.com/a/38330814/653632
	sel := r.newSelect().Columns(fmt.Sprintf("distinct rtrim(path, replace(path, '%s', ''))", string(os.PathSeparator))).
		Where(pathStartsWith(path))
	var res []string
	err := r.queryAll(sel, &res)
	return res, err
}

func (r *mediaFileRepository) deleteNotInPath(basePath string) error {
	path := cleanPath(basePath)

	// Set duplicate songs visible
	upd := Update(r.tableName).Set("visible", "T").Where(`
			(album_id, title, artist_id) in (select album_id, title, artist_id from media_file where substr(path, 1, length('` + path + `')) <> '` + path + `')
	`)
	c, err := r.executeSQL(upd)
	if err == nil {
		if c > 0 {
			log.Debug(r.ctx, "Update duplicate songs visible", "totalUpdated", c)
		}
	}

	sel := Delete(r.tableName).Where(NotEq(pathStartsWith(path)))
	c, err = r.executeSQL(sel)
	if err == nil {
		if c > 0 {
			log.Debug(r.ctx, "Deleted dangling tracks", "totalDeleted", c)
		}
	}
	return err
}

func (r *mediaFileRepository) Delete(id string) error {
	// Set duplicate songs visible
	upd := Update(r.tableName).Set("visible", "T").Where(`
			(album_id, title, artist_id) in (select album_id, title, artist_id from media_file where id = '` + id + `')
	`)
	c, err := r.executeSQL(upd)
	if err == nil {
		if c > 0 {
			log.Debug(r.ctx, "Update duplicate songs visible", "totalUpdated", c)
		}
	}
	return r.delete(Eq{"id": id})
}

// DeleteByPath delete from the DB all mediafiles that are direct children of path
func (r *mediaFileRepository) DeleteByPath(basePath string) (int64, error) {
	path := cleanPath(basePath)
	pathLen := utf8.RuneCountInString(path)

	// Set duplicate songs visible
	upd := Update(r.tableName).Set("visible", "T").Where(`
			(album_id, title, artist_id) in (select album_id, title, artist_id from media_file where substr(path, 1, length('` + path + `')) = '` + path + `')
	`)
	c, err := r.executeSQL(upd)
	if err == nil {
		if c > 0 {
			log.Debug(r.ctx, "Update duplicate songs visible", "totalUpdated", c)
		}
	}

	del := Delete(r.tableName).
		Where(And{pathStartsWith(path),
			Eq{fmt.Sprintf("substr(path, %d) glob '*%s*'", pathLen+2, string(os.PathSeparator)): 0}})
	log.Debug(r.ctx, "Deleting mediafiles by path", "path", path)
	return r.executeSQL(del)
}

func (r *mediaFileRepository) removeNonAlbumArtistIds() error {
	// upd := Update(r.tableName).Set("artist_id", "").Where(notExists("artist", ConcatExpr("id = artist_id")))
	// upd := Update(r.tableName).Set("artist_id", "").Where(notExists("artist", ConcatExpr("all_artist_ids like '%'||artist.id||'%'")))
	upd := `
		with media_file_artist as 
		(WITH RECURSIVE split(seq, art_id, art_id_str, id) AS (
					SELECT 0, '/', mf.artist_id||'/', mf.id from media_file mf
					UNION ALL SELECT
						seq+1,
						substr(art_id_str, 0, instr(art_id_str, '/')),
						substr(art_id_str, instr(art_id_str, '/')+1),
								id
					FROM split WHERE art_id_str != ''
				) SELECT split.id, count(a.id) as artist_count FROM split
				LEFT JOIN artist a on a.id = art_id
				where seq !=0 
				GROUP BY split.id)
		update media_file 
		set artist = case mfa.artist_count when 0 then '' else media_file.artist end
		from media_file_artist mfa
		where mfa.id = media_file.id
	`
	log.Debug(r.ctx, "Removing non-album artist_id")
	_, err := r.executeRawSQL(upd)
	return err
}

func (r *mediaFileRepository) ProcessDuplicateSongs() error {
	upd := Update(r.tableName).Set("visible", "F").Where(`
			(album_id, title, artist_id) in (
				select album_id, title, artist_id from media_file where visible='T' group by album_id, title, artist_id having count(*) > 1)
			and rowid not in (select min(rowid) from media_file where visible='T' group by album_id, title, artist_id having count(*) > 1)
			and visible='T'
	`)
	log.Debug(r.ctx, "Setting duplicate media file unshow")
	_, err := r.executeSQL(upd)
	return err
}

func (r *mediaFileRepository) Search(q string, offset int, size int) (model.MediaFiles, error) {
	results := model.MediaFiles{}
	err := r.doSearch(q, offset, size, &results, "title")
	return results, err
}

func (r *mediaFileRepository) Count(options ...rest.QueryOptions) (int64, error) {
	return r.CountAll(r.parseRestOptions(options...))
}

func (r *mediaFileRepository) Read(id string) (interface{}, error) {
	return r.Get(id)
}

func (r *mediaFileRepository) ReadAll(options ...rest.QueryOptions) (interface{}, error) {
	return r.GetAll(r.parseRestOptions(options...))
}

func (r *mediaFileRepository) EntityName() string {
	return "mediafile"
}

func (r *mediaFileRepository) NewInstance() interface{} {
	return &model.MediaFile{}
}

var _ model.MediaFileRepository = (*mediaFileRepository)(nil)
var _ model.ResourceRepository = (*mediaFileRepository)(nil)
