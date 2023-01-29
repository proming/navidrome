package persistence

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	. "github.com/Masterminds/squirrel"
	"github.com/beego/beego/v2/client/orm"
	"github.com/deluan/rest"
	"github.com/navidrome/navidrome/conf"
	"github.com/navidrome/navidrome/consts"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/utils"
)

type artistRepository struct {
	sqlRepository
	sqlRestful
	indexGroups utils.IndexGroups
}

type dbArtist struct {
	model.Artist   `structs:",flatten"`
	SimilarArtists string `structs:"similar_artists" json:"similarArtists"`
}

func NewArtistRepository(ctx context.Context, o orm.QueryExecutor) model.ArtistRepository {
	r := &artistRepository{}
	r.ctx = ctx
	r.ormer = o
	r.indexGroups = utils.ParseIndexGroups(conf.Server.IndexGroups)
	r.tableName = "artist"
	r.sortMappings = map[string]string{
		"name": "order_artist_name",
	}
	r.filterMappings = map[string]filterFunc{
		"id":      idFilter(r.tableName),
		"name":    fullTextFilter,
		"starred": booleanFilter,
	}
	return r
}

func (r *artistRepository) selectArtist(options ...model.QueryOptions) SelectBuilder {
	sql := r.newSelectWithAnnotation("artist.id", options...).Columns("artist.*")
	return r.withGenres(sql).GroupBy("artist.id")
}

func (r *artistRepository) CountAll(options ...model.QueryOptions) (int64, error) {
	sql := r.newSelectWithAnnotation("artist.id")
	sql = r.withGenres(sql)
	return r.count(sql, options...)
}

func (r *artistRepository) Exists(id string) (bool, error) {
	return r.exists(Select().Where(Eq{"artist.id": id}))
}

func (r *artistRepository) Put(a *model.Artist) error {
	a.FullText = getFullText(a.Name, utils.SplitAndJoinStrings(a.SortArtistName))
	dba := r.fromModel(a)
	_, err := r.put(dba.ID, dba)
	if err != nil {
		return err
	}
	if a.ID == consts.VariousArtistsID {
		return r.updateGenres(a.ID, r.tableName, nil)
	}
	return r.updateGenres(a.ID, r.tableName, a.Genres)
}

func (r *artistRepository) Get(id string) (*model.Artist, error) {
	sel := r.selectArtist().Where(Eq{"artist.id": id})
	var dba []dbArtist
	if err := r.queryAll(sel, &dba); err != nil {
		return nil, err
	}
	if len(dba) == 0 {
		return nil, model.ErrNotFound
	}
	res := r.toModels(dba)
	err := r.loadArtistGenres(&res)
	return &res[0], err
}

func (r *artistRepository) GetAll(options ...model.QueryOptions) (model.Artists, error) {
	sel := r.selectArtist(options...)
	var dba []dbArtist
	err := r.queryAll(sel, &dba)
	if err != nil {
		return nil, err
	}
	res := r.toModels(dba)
	err = r.loadArtistGenres(&res)
	return res, err
}

func (r *artistRepository) toModels(dba []dbArtist) model.Artists {
	res := model.Artists{}
	for i := range dba {
		a := dba[i]
		res = append(res, *r.toModel(&a))
	}
	return res
}

func (r *artistRepository) toModel(dba *dbArtist) *model.Artist {
	a := dba.Artist
	a.SimilarArtists = nil
	for _, s := range strings.Split(dba.SimilarArtists, ";") {
		fields := strings.Split(s, ":")
		if len(fields) != 2 {
			continue
		}
		name, _ := url.QueryUnescape(fields[1])
		a.SimilarArtists = append(a.SimilarArtists, model.Artist{
			ID:   fields[0],
			Name: name,
		})
	}
	return &a
}

func (r *artistRepository) fromModel(a *model.Artist) *dbArtist {
	dba := &dbArtist{Artist: *a}
	var sa []string

	for _, s := range a.SimilarArtists {
		sa = append(sa, fmt.Sprintf("%s:%s", s.ID, url.QueryEscape(s.Name)))
	}

	dba.SimilarArtists = strings.Join(sa, ";")
	return dba
}

func (r *artistRepository) getIndexKey(a *model.Artist) string {
	name := strings.ToLower(utils.NoArticle(a.Name))
	for k, v := range r.indexGroups {
		key := strings.ToLower(k)
		if strings.HasPrefix(name, key) {
			return v
		}
	}
	return "#"
}

// TODO Cache the index (recalculate when there are changes to the DB)
func (r *artistRepository) GetIndex() (model.ArtistIndexes, error) {
	all, err := r.GetAll(model.QueryOptions{Sort: "order_artist_name"})
	if err != nil {
		return nil, err
	}

	fullIdx := make(map[string]*model.ArtistIndex)
	for i := range all {
		a := all[i]
		ax := r.getIndexKey(&a)
		idx, ok := fullIdx[ax]
		if !ok {
			idx = &model.ArtistIndex{ID: ax}
			fullIdx[ax] = idx
		}
		idx.Artists = append(idx.Artists, a)
	}
	var result model.ArtistIndexes
	for _, idx := range fullIdx {
		result = append(result, *idx)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (r *artistRepository) Refresh(ids ...string) error {
	start := time.Now()
	type refreshArtist struct {
		model.Artist
		CurrentId string
		GenreIds  string
	}
	var artists []refreshArtist
	// sel := Select("f.album_artist_id as id", "f.album_artist as name", "count(*) as album_count", "a.id as current_id",
	// 	"group_concat(f.mbz_album_artist_id , ' ') as mbz_artist_id",
	// 	"f.sort_album_artist_name as sort_artist_name", "f.order_album_artist_name as order_artist_name",
	// 	"sum(f.song_count) as song_count", "sum(f.size) as size",
	// 	"alg.genre_ids").
	// 	From("album f").
	// 	LeftJoin("artist a on f.album_artist_id = a.id").
	// 	LeftJoin(`(select al.album_artist_id, group_concat(ag.genre_id, ' ') as genre_ids from album_genres ag
	// 			left join album al on al.id = ag.album_id where al.album_artist_id in ('` +
	// 		strings.Join(ids, "','") + `') group by al.album_artist_id) alg on alg.album_artist_id = f.album_artist_id`).
	// 	Where(Eq{"f.album_artist_id": ids}).
	// 	GroupBy("f.album_artist_id").OrderBy("f.id")
	sel := Select(`art.art_id as id, trim(art.art_name) as name, count(DISTINCT mf.album_id) as album_count,
	a.id as current_id, group_concat(mf.mbz_album_artist_id , ' ') as mbz_artist_id,
	trim(art.art_name) as sort_artist_name, 
	trim(art.art_name) as order_artist_name,
	COUNT(*) as song_count, sum(mf.size) as size,
	group_concat(mfg.genre_id, ' ') as genre_ids
	FROM media_file mf
	INNER JOIN (WITH RECURSIVE split(seq, art_name, art_name_str, art_id, art_id_str, id) AS (
		SELECT 0, '/', mf.artist||'/'||mf.album_artist||'/', '/', 
				mf.artist_id||'/'||mf.album_artist_id||'/', mf.id from media_file mf where mf.id in ('` + strings.Join(ids, "','") + `')
		UNION ALL SELECT
			seq+1,
			substr(art_name_str, 0, instr(art_name_str, '/')),
			substr(art_name_str, instr(art_name_str, '/')+1),
			substr(art_id_str, 0, instr(art_id_str, '/')),
			substr(art_id_str, instr(art_id_str, '/')+1),
					id
		FROM split WHERE art_name_str != ''
	) SELECT art_name, art_id, split.id FROM split 
  where seq !=0 
	group by art_name, art_id, split.id) art on mf.id = art.id
	LEFT JOIN artist a on art.art_id = a.id
	LEFT JOIN media_file_genres mfg on mf.id = mfg.media_file_id
	WHERE mf.id in ('` + strings.Join(ids, "','") + `') and mf.visible = 'T'
	GROUP BY art_id`)
	err := r.queryAll(sel, &artists)
	if err != nil {
		return err
	}

	toInsert := 0
	toUpdate := 0
	for _, ar := range artists {
		if ar.CurrentId != "" {
			toUpdate++
		} else {
			toInsert++
		}
		ar.MbzArtistID = getMostFrequentMbzID(r.ctx, ar.MbzArtistID, r.tableName, ar.Name)
		ar.Genres = getGenres(ar.GenreIds)
		ar.ExternalInfoUpdatedAt = time.Time{}
		err := r.Put(&ar.Artist)
		if err != nil {
			return err
		}
	}
	if toInsert > 0 {
		log.Debug(r.ctx, "Inserted new artists", "totalInserted", toInsert)
	}
	if toUpdate > 0 {
		log.Debug(r.ctx, "Updated artists", "totalUpdated", toUpdate)
	}

	log.Info("Artist refresh", "elapsedTime", time.Since(start))
	return err
}

func (r *artistRepository) purgeEmpty() error {
	// del := Delete(r.tableName).Where("id not in (select distinct(album_artist_id) from album)")
	// del := Delete(r.tableName).Where(notExists("media_file", ConcatExpr("all_artist_ids like '%'||artist.id||'%'")))
	del := `
	with media_file_artist as 
	(WITH RECURSIVE split(seq, art_id, art_id_str, id) AS (
		SELECT 0, ' ', mf.all_artist_ids||' ', mf.id from media_file mf
		UNION ALL SELECT
			seq+1,
			substr(art_id_str, 0, instr(art_id_str, ' ')),
			substr(art_id_str, instr(art_id_str, ' ')+1),
					id
		FROM split WHERE art_id_str != ''
	) SELECT art_id, count(split.id) as media_count FROM split
	where seq !=0 
	GROUP BY art_id)
	delete from artist 
	where EXISTS(select 1 from media_file_artist mfa where mfa.art_id = artist.id and media_count=0)
	`

	c, err := r.executeRawSQL(del)
	if err == nil {
		if c > 0 {
			log.Debug(r.ctx, "Purged empty artists", "totalDeleted", c)
		}
	}
	return err
}

func (r *artistRepository) updateAlbumCount() error {
	upd := `
	with artist_album_count as 
	(WITH RECURSIVE split(seq, art_id, art_id_str, id) AS (
				SELECT 0, '/', mf.album_artist_id||'/', mf.id from album mf
				UNION ALL SELECT
					seq+1,
					substr(art_id_str, 0, instr(art_id_str, '/')),
					substr(art_id_str, instr(art_id_str, '/')+1),
							id
				FROM split WHERE art_id_str != ''
			) SELECT art_id, count(split.id) as album_count FROM split 
			where seq !=0 
			group by art_id)
	update artist 
	set album_count = IFNULL(abc.album_count, 0)
	from artist_album_count abc
	where abc.art_id = artist.id
	`
	c, err := r.executeRawSQL(upd)
	if err == nil {
		if c > 0 {
			log.Debug(r.ctx, "Update artist album count", "totalUpdated", c)
		}
	} else {
		return err
	}

	upd = `
	with artist_song_count as 
	(WITH RECURSIVE split(seq, art_id, art_id_str, id, size) AS (
		SELECT 0, ' ', mf.all_artist_ids||' ', mf.id, mf.size from media_file mf
		UNION ALL SELECT
				seq+1,
				substr(art_id_str, 0, instr(art_id_str, ' ')),
				substr(art_id_str, instr(art_id_str, ' ')+1),
				id,
				size
		FROM split WHERE art_id_str != ''
		) select art_id, count(id) as song_count, sum(size) as size from (SELECT art_id, id, size FROM split 
		where seq !=0 
		GROUP BY art_id, id)
		GROUP BY art_id
	)
	update artist 
	set song_count = IFNULL(abc.song_count, 0), size=IFNULL(abc.size, 0)
	from artist_song_count abc
	where abc.art_id = artist.id
	`
	c, err = r.executeRawSQL(upd)
	if err == nil {
		if c > 0 {
			log.Debug(r.ctx, "Update artist song count", "totalUpdated", c)
		}
	}

	return err
}

func (r *artistRepository) Search(q string, offset int, size int) (model.Artists, error) {
	var dba []dbArtist
	err := r.doSearch(q, offset, size, &dba, "name")
	if err != nil {
		return nil, err
	}
	return r.toModels(dba), nil
}

func (r *artistRepository) Count(options ...rest.QueryOptions) (int64, error) {
	return r.CountAll(r.parseRestOptions(options...))
}

func (r *artistRepository) Read(id string) (interface{}, error) {
	return r.Get(id)
}

func (r *artistRepository) ReadAll(options ...rest.QueryOptions) (interface{}, error) {
	return r.GetAll(r.parseRestOptions(options...))
}

func (r *artistRepository) EntityName() string {
	return "artist"
}

func (r *artistRepository) NewInstance() interface{} {
	return &model.Artist{}
}

var _ model.ArtistRepository = (*artistRepository)(nil)
var _ model.ResourceRepository = (*artistRepository)(nil)
