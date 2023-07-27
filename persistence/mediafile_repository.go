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
		"artist": "order_artist_name asc, order_album_name asc, release_date asc, disc_number asc, track_number asc",
		"album":  "order_album_name asc, release_date asc, disc_number asc, track_number asc, order_artist_name asc, title asc",
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
	sql := r.newSelectWithAnnotation("media_file.id")
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
	_, err := r.put(m.ID, m)
	if err != nil {
		return err
	}
	err = r.updateGenres(m.ID, r.tableName, m.Genres)
	if err != nil {
		return err
	}
	return r.updateMediaFileArtist(m.ID, r.tableName, fmt.Sprintf("%s/%s", m.ArtistID, m.AlbumArtistID), fmt.Sprintf("%s/%s", m.Artist, m.AlbumArtist))
}

func (r *mediaFileRepository) updateMediaFileArtist(id string, tableName string, artists string, artistNames string) error {
	del := Delete(tableName + "_artist_list").Where(Eq{tableName + "_id": id})
	_, err := r.executeSQL(del)
	if err != nil {
		return err
	}

	if len(artists) == 0 {
		return nil
	}
	artistSplit := strings.Split(artists, "/")
	nameSplit := strings.Split(artistNames, "/")
	artistMap := make(map[string]string)
	for idx, a := range artistSplit {
		artistMap[a] = nameSplit[idx]
	}

	ins := Insert(tableName+"_artist_list").Columns("artist_id", "artist_name", tableName+"_id")
	for k, v := range artistMap {
		ins = ins.Values(k, v, id)
	}

	_, err = r.executeSQL(ins)
	return err
}

func (r *mediaFileRepository) selectMediaFile(options ...model.QueryOptions) SelectBuilder {
	sql := r.newSelectWithAnnotation("media_file.id", options...).Columns("media_file.*")
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
	sq := r.selectMediaFile(options...)
	res := model.MediaFiles{}
	err := r.queryAll(sq, &res)
	if err != nil {
		return nil, err
	}
	err = r.loadMediaFileGenres(&res)
	return res, err
}

func (r *mediaFileRepository) QueryAll(ids []string, response interface{}) error {
	sel := Select(`
	mf.id, mf.path, mf.title, mf.album, mfa.artist_name artist, mfa.artist_id, mf.album_artist, mf.album_id, 
	mf.has_cover_art, mf.track_number, mf.disc_number, mf.year, mf.size, mf.suffix, mf.duration, 
	mf.bit_rate, mf.genre, mf.compilation, mf.created_at, mf.updated_at, mf.full_text, mf.album_artist_id, 
	mf.order_album_name, mf.order_album_artist_name, mf.order_artist_name, mf.sort_album_name, 
	mf.sort_artist_name, mf.sort_album_artist_name, mf.sort_title, mf.disc_subtitle, mf.mbz_recording_id, 
	mf.mbz_album_id, mf.mbz_artist_id, mf.mbz_album_artist_id, mf.mbz_album_type, mf.mbz_album_comment, 
	mf.catalog_num, mf.comment, mf.lyrics, mf.bpm, mf.channels, mf.order_title, mf.mbz_release_track_id, 
	mf.rg_album_gain, mf.rg_album_peak, mf.rg_track_gain, mf.rg_track_peak, mf.all_artist_ids
	`).From("media_file mf").LeftJoin("media_file_artist_list mfa on mf.id=mfa.media_file_id").Where(Eq{"mfa.artist_id": ids})
	return r.queryAll(sel, response)
}

func (r *mediaFileRepository) FindByPath(path string) (*model.MediaFile, error) {
	sel := r.newSelect().Columns("*").Where(Like{"path": path})
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
	sel := Delete(r.tableName).Where(NotEq(pathStartsWith(path)))
	c, err := r.executeSQL(sel)
	if err == nil {
		if c > 0 {
			log.Debug(r.ctx, "Deleted dangling tracks", "totalDeleted", c)
		}
	}
	return err
}

func (r *mediaFileRepository) Delete(id string) error {
	return r.delete(Eq{"id": id})
}

// DeleteByPath delete from the DB all mediafiles that are direct children of path
func (r *mediaFileRepository) DeleteByPath(basePath string) (int64, error) {
	path := cleanPath(basePath)
	pathLen := utf8.RuneCountInString(path)
	del := Delete(r.tableName).
		Where(And{pathStartsWith(path),
			Eq{fmt.Sprintf("substr(path, %d) glob '*%s*'", pathLen+2, string(os.PathSeparator)): 0}})
	log.Debug(r.ctx, "Deleting mediafiles by path", "path", path)
	return r.executeSQL(del)
}

func (r *mediaFileRepository) removeNonAlbumArtistIds() error {
	// upd := Update(r.tableName).Set("artist_id", "").Where(notExists("artist", ConcatExpr("id = artist_id")))
	upd := Update(r.tableName).Set("artist_id", "").Where("id not in (select distinct(media_file_id) from media_file_artist_list)")
	log.Debug(r.ctx, "Removing non-album artist_ids")
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
