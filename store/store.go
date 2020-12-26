package store

// TODO:
// - all reads to be performed via Iter
// - all writes to be performed by Batch
// - kill all the direct-manipulation functions!

import (
	"database/sql"
	"fmt"
	"sort"
	//"github.com/bcampbell/arts/arts"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type ProgressFn func(expected int, completed int)

//
var defaultField string = "content"

// indexer is the interface for maintaining a fulltext index on the database
type indexer interface {
	// add or update articles
	index(...*Article) error
	zap(...ArtID) error
	search(string, string) (ArtList, error)
	Close() error
}

// Store is the core representation of our data set.
// Provides base methods for querying, tagging, whatever
type Store struct {
	db   *sql.DB
	dbug Logger
	idx  indexer
	// lang is default language used for indexing
	lang   string
	loc    *time.Location
	dbFile string
}

/* TEMP - cheesy access to db directly */
func (store *Store) DB() *sql.DB {
	return store.db
}

// Open an existing database & index. Returns error if non-existant.
func Open(dbFile string, dbug Logger, defaultLang string, loc *time.Location) (*Store, error) {
	_, err := os.Stat(dbFile)
	if err != nil {
		//if os.IsNotExist(err) { ... }
		return nil, err
	}
	return internalNew(dbFile, dbug, defaultLang, loc, true)
}

// Create creates a new database and index. Returns an error if already existant.
func Create(dbFile string, dbug Logger, defaultLang string, loc *time.Location) (*Store, error) {
	_, err := os.Stat(dbFile)
	if os.IsNotExist(err) {
		return internalNew(dbFile, dbug, defaultLang, loc, true)
	} else {
		return nil, fmt.Errorf("%s already exists", dbFile)
	}
}

// DEPRECATED - use Open() or Create() instead
// defaultLang is lang used when creating a new db (or updating an older version).
// Otherwise Lang set from existing DB.
func New(dbFile string, dbug Logger, defaultLang string, loc *time.Location) (*Store, error) {
	return internalNew(dbFile, dbug, defaultLang, loc, true)
}

// NewWithoutIndex opens a store without a fulltext interface
func NewWithoutIndex(dbFile string, dbug Logger, defaultLang string, loc *time.Location) (*Store, error) {
	return internalNew(dbFile, dbug, defaultLang, loc, false)
}

func internalNew(dbFile string, dbug Logger, defaultLang string, loc *time.Location, useBleve bool) (*Store, error) {
	store := &Store{dbug: dbug, lang: defaultLang, loc: loc, dbFile: dbFile}

	var err error

	//
	store.db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	// make sure schema is current
	err = store.ensureSchema()
	if err != nil {
		return nil, err
	}

	// update the lang from the db itself.
	err = store.db.QueryRow(`SELECT v FROM settings WHERE n='lang'`).Scan(&store.lang)
	if err != nil {
		return nil, err
	}

	err = store.cleanDanglingData()
	if err != nil {
		return nil, err
	}

	if useBleve {
		indexDir := dbFile + ".bleve"
		fi, err := os.Stat(indexDir)
		if os.IsNotExist(err) {
			store.dbug.Printf("Create new index from scratch\n")
			// new indexer
			store.idx, err = newBleveIndex(store.dbug, indexDir, store.lang, loc)
			if err != nil {
				return nil, err
			}

			// index all articles
			allArts, err := store.AllArts()
			if err != nil {
				return nil, err
			}
			err = store.reindex(allArts)
			if err != nil {
				return nil, err
			}

		} else {
			store.dbug.Printf("Open existing index %s\n", indexDir)
			// open existing indexer
			if !fi.IsDir() {
				return nil, fmt.Errorf("expected %s to be a directory", indexDir)
			}
			store.idx, err = openBleveIndex(store.dbug, indexDir, loc)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", indexDir, err)
			}
		}
	} else {
		store.idx = &dummyIndex{}
	}
	//

	return store, nil
}

// Filename returns the filename of the .db file underlying this store
func (store *Store) Filename() string {
	return store.dbFile
}

func (store *Store) Close() {
	if store.db != nil {
		//store.dbug.Printf("Close sqlite db\n")
		store.db.Close()
		store.db = nil
	}

	if store.idx != nil {
		err := store.idx.Close()
		if err != nil {
			store.dbug.Printf("ERROR closing index: %s\n", err)
		}
		store.idx = nil
	}
}

func (store *Store) TotalArts() int {
	var cnt int
	var err error
	err = store.db.QueryRow(`SELECT COUNT(*) FROM article`).Scan(&cnt)
	if err != nil {
		store.dbug.Printf("TotalArts() failed: %s\n", err)
		return 0
	}

	return cnt
}

// tidy up some potentially-dangling data (due to faulty delete mechanism!)
func (store *Store) cleanDanglingData() error {
	db := store.db

	// clean up article_author table
	result, err := db.Exec("DELETE FROM article_author WHERE article_id NOT IN (SELECT id FROM article)")
	if err != nil {
		return err
	}

	var n int64
	n, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if n > 0 {
		store.dbug.Printf("WARNING: cleaned up %d dangling author entries\n", n)
	}

	// clean up article_link table
	result, err = db.Exec("DELETE FROM article_link WHERE article_id NOT IN (SELECT id FROM article)")
	if err != nil {
		return err
	}
	n, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if n > 0 {
		store.dbug.Printf("WARNING: cleaned up %d dangling article_link entries\n", n)
	}

	// clean up article_keyword table
	result, err = db.Exec("DELETE FROM article_keyword WHERE article_id NOT IN (SELECT id FROM article)")
	if err != nil {
		return err
	}
	n, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if n > 0 {
		store.dbug.Printf("WARNING: cleaned up %d dangling article_keyword entries\n", n)
	}

	return nil
}

// read in articles from DB
/*
func (store *Store) readAllArts() (ArtList, error) {
	db := store.db

	tab := map[ArtID]*Article{}

	// now grab all the articles!
	rows, err := db.Query("SELECT id,canonical_url,headline,content,published,updated,pub,section,retweet_count,favourite_count FROM article")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		a := &Article{}
		err = rows.Scan(&a.ID, &a.CanonicalURL, &a.Headline, &a.Content, &a.Published, &a.Updated, &a.Pub, &a.Section, &a.Retweets, &a.Favourites)
		if err != nil {
			return nil, err
		}
		tab[a.ID] = a
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// URLs
	rows, err = db.Query("SELECT article_id,url FROM article_url")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID ArtID
		var u string
		err = rows.Scan(&artID, &u)
		if err != nil {
			return nil, err
		}
		art := tab[artID]
		art.URLs = append(art.URLs, u)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// Keywords
	rows, err = db.Query("SELECT article_id,name,url FROM article_keyword")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID ArtID
		// TODO: restore use of full Keyword struct
		var name string
		var u string
		err = rows.Scan(&artID, &name, &u)
		if err != nil {
			return nil, err
		}
		art := tab[artID]
		art.Keywords = append(art.Keywords, name)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// Links
	rows, err = db.Query("SELECT article_id,url FROM article_link")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID ArtID
		var link string
		err = rows.Scan(&artID, &link)
		if err != nil {
			return nil, err
		}
		art := tab[artID]
		art.Links = append(art.Links, link)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// tags
	rows, err = db.Query("SELECT article_id,tag FROM article_tag")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID ArtID
		var tag string
		err = rows.Scan(&artID, &tag)
		if err != nil {
			return nil, err
		}
		art := tab[artID]
		art.Tags = append(art.Tags, tag)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// authors
	danglingAuthorCnt := 0
	rows, err = db.Query("SELECT article_id,name FROM article_author")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID ArtID
		var name string
		err = rows.Scan(&artID, &name)
		if err != nil {
			return nil, err
		}
		art, got := tab[artID]
		if got {
			art.Authors = append(art.Authors, Author{Name: name})
		} else {
			danglingAuthorCnt++

		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if danglingAuthorCnt > 0 {
		store.dbug.Printf("WARNING: database has %d dangling authors from deleted articles.\n", danglingAuthorCnt)
		store.dbug.Printf("not a big deal, but db can be repaired manually via sql:\n")
		store.dbug.Printf("  DELETE FROM article_author WHERE article_id NOT IN (SELECT id FROM article);\n")
	}

	// all done
	out := ArtList{}
	for _, art := range tab {
		// evil hack (TODO: less evil, please)
		art.Byline = art.BylineString()

		out = append(out, art.ID)
	}
	return out, nil
}
*/

// standin - return IDs of all articles
func (store *Store) AllArts() (ArtList, error) {

	rows, err := store.db.Query("SELECT id FROM article")
	if err != nil {
		return nil, err
	}
	out := make(ArtList, 0, 2000)
	for rows.Next() {
		var artID ArtID
		err = rows.Scan(&artID)
		if err != nil {
			return nil, err
		}
		out = append(out, artID)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return out, nil
}

// search performs a search and returns the results
// XYZZY: add sort criteria
func (store *Store) Search(queryString string) (ArtList, error) {
	// cheesy-ass hackery
	if queryString == "" {
		return store.AllArts()
	}

	out, err := store.idx.search(queryString, "")
	if err != nil {
		store.dbug.Printf("Search(%s) error: %s\n", queryString, err)
		return nil, err
	}
	//store.dbug.Printf("Search(%s): %d matches\n", queryString, len(out))

	return out, nil
}

func fileNameFromQuery(q string) string {
	colon := regexp.MustCompile(`:\s*`)
	spc := regexp.MustCompile(`\s+`)
	chars := regexp.MustCompile(`[^-\w]`)
	f := q
	f = colon.ReplaceAllString(f, "-")
	f = spc.ReplaceAllString(f, "_")
	f = chars.ReplaceAllString(f, "")
	return f
}

func (store *Store) AddTags(arts ArtList, tags []string) (ArtList, error) {

	if store.db == nil {
		return ArtList{}, nil
	}

	begun := time.Now()
	// apply to db
	tx, err := store.db.Begin()
	if err != nil {
		return nil, err
	}

	affected := ArtList{}
	for _, tag := range tags {
		newlyTagged, err := store.addTag(tx, arts, tag)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		affected = affected.Union(newlyTagged)
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	elapsedDB := time.Now().Sub(begun)

	begun = time.Now()
	// now apply to belve index
	// a bit cumbersome, but bleve has no way to update select fields?
	err = store.reindex(affected)
	if err != nil {
		return nil, err
	}
	elapsedBleve := time.Now().Sub(begun)

	store.dbug.Printf("add tags %q: affected %d articles (%s db, %s bleve)\n", tags, len(affected), elapsedDB, elapsedBleve)

	return affected, nil
}

func (store *Store) reindex(arts ArtList) error {

	start := 0
	for start < len(arts) {
		n := 200
		end := start + n
		if end > len(arts) {
			end = len(arts)
		}

		//store.dbug.Printf("chunk %d:%d (of %d)\n", start, end, len(arts))
		chunk := arts[start:end]
		start = end

		fullArts, err := store.Fetch(chunk...)
		if err != nil {
			return err
		}

		err = store.idx.index(fullArts...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (store *Store) addTag(tx *sql.Tx, arts ArtList, tag string) (ArtList, error) {

	tag = strings.ToLower(tag)

	// find any article which already have the tag
	gotRows, err := store.db.Query(`SELECT article_id FROM article_tag WHERE tag=? AND article_id IN (`+arts.StringList()+`)`, tag)
	if err != nil {
		return nil, err
	}

	got := ArtList{}
	for gotRows.Next() {
		var id ArtID
		err = gotRows.Scan(&id)
		if err != nil {
			return nil, err
		}
		got = append(got, id)
	}

	// tag the ones that need it
	affected := arts.Subtract(got)
	insStmt, err := tx.Prepare("INSERT INTO article_tag(article_id,tag) VALUES(?,?)")
	if err != nil {
		return nil, err
	}
	defer insStmt.Close()

	for _, artID := range affected {
		_, err = insStmt.Exec(artID, tag)
		if err != nil {
			return nil, err
		}
	}

	return affected, nil
}

func (store *Store) RemoveTags(arts ArtList, tags []string) (ArtList, error) {

	if store.db == nil {
		return ArtList{}, nil
	}

	// apply to db
	tx, err := store.db.Begin()
	if err != nil {
		return nil, err
	}

	affected := ArtList{}
	for _, tag := range tags {
		newly, err := store.removeTag(tx, arts, tag)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		store.dbug.Printf("removed '%s' tag from %d articles\n", tag, len(newly))
		affected = affected.Union(newly)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// now apply to belve index
	// a bit cumbersome, but bleve has no way to update select fields?
	err = store.reindex(affected)
	if err != nil {
		return nil, err
	}
	return affected, nil
}

func (store *Store) removeTag(tx *sql.Tx, arts ArtList, tag string) (ArtList, error) {
	tag = strings.ToLower(tag)

	// find which articles need the tag removed
	rows, err := store.db.Query(`SELECT article_id FROM article_tag WHERE tag=? AND article_id IN (`+idList(arts)+`)`, tag)

	if err != nil {
		return nil, err
	}

	got := ArtList{}
	for rows.Next() {
		var id ArtID
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		got = append(got, id)
	}

	_, err = store.db.Exec(`DELETE FROM article_tag WHERE tag=? AND article_id IN (`+idList(got)+`)`, tag)
	if err != nil {
		return nil, err
	}

	return got, nil
}

// Delete articles from the store (and index)
// if progress is non-nil, it'll be called at regularish intervals
func (store *Store) Delete(arts ArtList, progress ProgressFn) error {
	// delete from bleve index
	begun := time.Now()
	store.dbug.Printf("delete from index\n")
	err := store.idx.zap(arts...)
	if err != nil {
		return err
	}
	elapsedBleve := time.Now().Sub(begun)

	// now delete from db
	store.dbug.Printf("delete from db\n")
	begun = time.Now()
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}

	affected, err := store.doDelete(tx, arts, progress)
	if err != nil {
		store.dbug.Printf("error, rolling back\n")
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	elapsedDB := time.Now().Sub(begun)
	store.dbug.Printf("Delete affected %d articles (%s in db, %s in bleve)\n", affected, elapsedDB, elapsedBleve)
	return nil
}

func (store *Store) doDelete(tx *sql.Tx, arts ArtList, progress ProgressFn) (int64, error) {

	var affected int64 = 0
	// TODO: maybe use "on delete cascade" to let the db handle the details...
	// (requires an sqlite pragma to enable foreign keys)
	delTagsStmt, err := tx.Prepare("DELETE FROM article_tag WHERE article_id=?")
	if err != nil {
		return 0, err
	}
	defer delTagsStmt.Close()

	delURLsStmt, err := tx.Prepare("DELETE FROM article_url WHERE article_id=?")
	if err != nil {
		return 0, err
	}
	defer delURLsStmt.Close()

	delAuthorsStmt, err := tx.Prepare("DELETE FROM article_author WHERE article_id=?")
	if err != nil {
		return 0, err
	}
	defer delAuthorsStmt.Close()

	delKeywordsStmt, err := tx.Prepare("DELETE FROM article_keyword WHERE article_id=?")
	if err != nil {
		return 0, err
	}
	defer delKeywordsStmt.Close()

	delLinksStmt, err := tx.Prepare("DELETE FROM article_link WHERE article_id=?")
	if err != nil {
		return 0, err
	}
	defer delLinksStmt.Close()

	delArtStmt, err := tx.Prepare("DELETE FROM article WHERE id=?")
	if err != nil {
		return 0, err
	}
	defer delArtStmt.Close()

	for idx, artID := range arts {

		_, err = delTagsStmt.Exec(artID)
		if err != nil {
			return 0, err
		}

		_, err = delURLsStmt.Exec(artID)
		if err != nil {
			return 0, err
		}

		_, err = delLinksStmt.Exec(artID)
		if err != nil {
			return 0, err
		}

		_, err = delKeywordsStmt.Exec(artID)
		if err != nil {
			return 0, err
		}

		_, err = delAuthorsStmt.Exec(artID)
		if err != nil {
			return 0, err
		}

		r, err := delArtStmt.Exec(artID)
		if err != nil {
			return 0, err
		}
		// check we actually deleted something...
		foo, err := r.RowsAffected()
		if err != nil {
			return 0, err
		}
		affected += foo

		if progress != nil {
			progress(len(arts), idx)
		}
	}

	return affected, nil
}

func (store *Store) FindArt(urls []string) (ArtID, error) {
	placeholders := make([]string, len(urls))
	params := make([]interface{}, len(urls))
	for i, _ := range urls {
		placeholders[i] = "?"
		params[i] = urls[i]
	}
	foo := fmt.Sprintf(`SELECT article_id FROM article_url WHERE url IN(%s)`, strings.Join(placeholders, ","))

	var artID ArtID
	err := store.db.QueryRow(foo, params...).Scan(&artID)
	if err == sql.ErrNoRows {
		return 0, nil // article not found
	} else if err != nil {
		return 0, err
	} else {
		return artID, nil
	}
}

func (store *Store) Stash(arts []*Article) error {
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}

	for _, art := range arts {
		if art.ID != 0 {
			panic("article id already set")
		}

		err = store.doStash(tx, art)
		if err != nil {
			break
		}
	}

	if err == nil {
		// update the index
		err = store.idx.index(arts...)
	}

	if err == nil {
		err = tx.Commit()
	} else {
		err2 := tx.Rollback()
		if err2 != nil {
			err = err2
		}
	}
	return err
}

func (store *Store) doStash(tx *sql.Tx, art *Article) error {
	var result sql.Result
	result, err := tx.Exec(`INSERT INTO article(canonical_url, headline, content,published,updated,pub,section,retweet_count,favourite_count)
        values(?,?,?,?,?,?,?,?,?)`,
		art.CanonicalURL,
		art.Headline,
		art.Content,
		art.Published,
		art.Updated,
		art.Pub,
		art.Section,
		art.Retweets,
		art.Favourites)
	if err != nil {
		return err
	}

	if artID, err := result.LastInsertId(); err == nil {
		art.ID = ArtID(artID)
	} else {
		return err
	}

	// add urls
	urlStmt, err := tx.Prepare("INSERT INTO article_url(article_id,url) VALUES(?,?)")
	if err != nil {
		return err
	}
	defer urlStmt.Close()

	for _, u := range art.URLs {
		_, err = urlStmt.Exec(art.ID, u)
		if err != nil {
			return err
		}
	}

	// add tags
	tagStmt, err := tx.Prepare("INSERT INTO article_tag(article_id,tag) VALUES(?,?)")
	if err != nil {
		return err
	}
	defer tagStmt.Close()
	for _, tag := range art.Tags {
		_, err = tagStmt.Exec(art.ID, tag)
		if err != nil {
			return err
		}
	}
	// add authors (TODO: other author fields)
	authorStmt, err := tx.Prepare("INSERT INTO article_author(article_id,name) VALUES(?,?)")
	if err != nil {
		return err
	}
	defer authorStmt.Close()
	for _, author := range art.Authors {
		_, err = authorStmt.Exec(art.ID, author.Name)
		if err != nil {
			return err
		}
	}

	// add keywords
	kwStmt, err := tx.Prepare("INSERT INTO article_keyword(article_id,name,url) VALUES(?,?,?)")
	if err != nil {
		return err
	}
	defer kwStmt.Close()

	for _, kw := range art.Keywords {
		// TODO: restore use of full keyword struct
		_, err = kwStmt.Exec(art.ID, kw, "")
		if err != nil {
			return err
		}
	}

	// add resolved links
	linkStmt, err := tx.Prepare("INSERT INTO article_link(article_id,url) VALUES(?,?)")
	if err != nil {
		return err
	}
	defer linkStmt.Close()

	for _, link := range art.Links {
		_, err = linkStmt.Exec(art.ID, link)
		if err != nil {
			return err
		}
	}

	// add similar articles
	simStmt, err := tx.Prepare("INSERT INTO similar(article_id, other_id, tag) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer simStmt.Close()
	for _, match := range art.Similar {
		_, err = tagStmt.Exec(art.ID, match.ID, match.Score)
		if err != nil {
			return err
		}
	}

	return nil
}

// HACK - update all the links for the given articles
func (store *Store) UpdateLinks(arts ArtList) error {
	/* XYZZY */
	return nil
	/*
		tx, err := store.db.Begin()
		if err != nil {
			return err
		}

		delStmt, err := tx.Prepare("DELETE FROM article_link WHERE article_id=?")
		if err != nil {
			return err
		}
		defer delStmt.Close()

		insStmt, err := tx.Prepare("INSERT INTO article_link(article_id,url) VALUES(?,?)")
		if err != nil {
			return err
		}
		defer insStmt.Close()

		for _, art := range arts {
			// zap old links
			_, err = delStmt.Exec(art.ID)
			if err != nil {
				return err
			}

			// add new links
			for _, link := range art.Links {
				_, err = insStmt.Exec(art.ID, link)
				if err != nil {
					return err
				}
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		return nil
	*/
}

func idList(ids []ArtID) string {
	frags := make([]string, len(ids))
	for idx, id := range ids {
		frags[idx] = strconv.Itoa(int(id))
	}
	return strings.Join(frags, ",")

}

func (store *Store) Fetch(artIDs ...ArtID) ([]*Article, error) {
	if len(artIDs) > 0 {
		return store.doFetch(artIDs)
	}
	return []*Article{}, nil
}

func (store *Store) FetchAll() ([]*Article, error) {
	return store.doFetch(ArtList{})
}

// read in articles from DB. Empty list means fetch _all_ arts :-)
func (store *Store) doFetch(artIDs ArtList) ([]*Article, error) {
	db := store.db

	tab := map[ArtID]*Article{}

	var where string
	if len(artIDs) > 0 {
		where = " WHERE id IN (" + idList(artIDs) + ")"
	}
	// now grab all the articles!
	rows, err := db.Query(`SELECT id,canonical_url,headline,content,published,updated,pub,section,retweet_count,favourite_count FROM article` + where)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		a := &Article{}
		err = rows.Scan(&a.ID, &a.CanonicalURL, &a.Headline, &a.Content, &a.Published, &a.Updated, &a.Pub, &a.Section, &a.Retweets, &a.Favourites)
		if err != nil {
			return nil, err
		}
		tab[a.ID] = a
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if len(artIDs) > 0 {
		// all auxillary tables indexed by article_id
		where = " WHERE article_id IN (" + idList(artIDs) + ")"
	}

	// URLs
	rows, err = db.Query("SELECT article_id,url FROM article_url" + where)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID ArtID
		var u string
		err = rows.Scan(&artID, &u)
		if err != nil {
			return nil, err
		}
		art := tab[artID]
		art.URLs = append(art.URLs, u)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// Keywords
	rows, err = db.Query("SELECT article_id,name,url FROM article_keyword" + where)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID ArtID
		// TODO: restore use of full Keyword struct
		var name string
		var u string
		err = rows.Scan(&artID, &name, &u)
		if err != nil {
			return nil, err
		}
		art := tab[artID]
		art.Keywords = append(art.Keywords, name)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// Links
	rows, err = db.Query("SELECT article_id,url FROM article_link" + where)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID ArtID
		var link string
		err = rows.Scan(&artID, &link)
		if err != nil {
			return nil, err
		}
		art := tab[artID]
		art.Links = append(art.Links, link)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// tags
	rows, err = db.Query("SELECT article_id,tag FROM article_tag" + where)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID ArtID
		var tag string
		err = rows.Scan(&artID, &tag)
		if err != nil {
			return nil, err
		}
		art := tab[artID]
		art.Tags = append(art.Tags, tag)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// authors
	danglingAuthorCnt := 0
	rows, err = db.Query("SELECT article_id,name FROM article_author" + where)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID ArtID
		var name string
		err = rows.Scan(&artID, &name)
		if err != nil {
			return nil, err
		}
		art, got := tab[artID]
		if got {
			art.Authors = append(art.Authors, Author{Name: name})
		} else {
			danglingAuthorCnt++

		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// Similar articles
	rows, err = db.Query("SELECT article_id, other_id, score FROM similar" + where)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var match Match
		var artID ArtID
		err = rows.Scan(&artID, &match.ID, &match.Score)
		if err != nil {
			return nil, err
		}
		art := tab[artID]
		art.Similar = append(art.Similar, match)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if danglingAuthorCnt > 0 {
		store.dbug.Printf("WARNING: found %d dangling authors from deleted articles.\n", danglingAuthorCnt)
		store.dbug.Printf("not a big deal, but db can be repaired manually via sql:\n")
		store.dbug.Printf("  DELETE FROM article_author WHERE article_id NOT IN (SELECT id FROM article);\n")
	}

	// all done
	out := []*Article{}
	for _, art := range tab {

		sort.Strings(art.Tags)
		sort.Strings(art.Keywords)
		sort.Strings(art.Links)

		// evil hack (TODO: less evil, please)
		art.Byline = art.BylineString()
		out = append(out, art)
	}
	return out, nil
}

type SortOrder int

const (
	Ascending SortOrder = iota
	Descending
)

// Sort an article list (by hitting the sqlite db)
func (store *Store) Sort(artIDs ArtList, fieldName string, order SortOrder) (ArtList, error) {

	dbOrder := "ASC"
	if order == Descending {
		dbOrder = "DESC"
	}

	q := ""
	dbField := ""
	switch fieldName {
	case "headline":
		dbField = "headline"
		q = fmt.Sprintf("SELECT id FROM article WHERE id IN (%s) ORDER BY %s %s", artIDs.StringList(), dbField, dbOrder)
	case "published":
		dbField = "published"
		q = fmt.Sprintf("SELECT id FROM article WHERE id IN (%s) ORDER BY %s %s", artIDs.StringList(), dbField, dbOrder)
	case "pub":
		dbField = "pub"
		q = fmt.Sprintf("SELECT id FROM article WHERE id IN (%s) ORDER BY %s %s", artIDs.StringList(), dbField, dbOrder)
	case "section":
		dbField = "section"
		q = fmt.Sprintf("SELECT id FROM article WHERE id IN (%s) ORDER BY %s %s", artIDs.StringList(), dbField, dbOrder)
	case "retweets":
		dbField = "retweet_count"
		q = fmt.Sprintf("SELECT id FROM article WHERE id IN (%s) ORDER BY %s %s", artIDs.StringList(), dbField, dbOrder)
	case "favourites":
		dbField = "favourite_count"
		q = fmt.Sprintf("SELECT id FROM article WHERE id IN (%s) ORDER BY %s %s", artIDs.StringList(), dbField, dbOrder)
	case "url":
		q = fmt.Sprintf(
			`SELECT DISTINCT a.id
            FROM (article a LEFT JOIN article_link l ON l.article_id=a.id)
            WHERE a.id IN (%s) ORDER BY canonical_url %s`, artIDs.StringList(), dbOrder)
	case "tags":
		q = fmt.Sprintf(
			`SELECT DISTINCT a.id
            FROM (article a LEFT JOIN article_tag t ON t.article_id=a.id)
            WHERE a.id IN (%s) ORDER BY t.tag %s`, artIDs.StringList(), dbOrder)
	case "keywords":
		q = fmt.Sprintf(
			`SELECT DISTINCT a.id
            FROM (article a LEFT JOIN article_keyword kw ON kw.article_id=a.id)
            WHERE a.id IN (%s) ORDER BY kw.name %s`, artIDs.StringList(), dbOrder)
	case "byline":
		q = fmt.Sprintf(
			`SELECT DISTINCT a.id
            FROM (article a LEFT JOIN article_author auth ON auth.article_id=a.id)
            WHERE a.id IN (%s) ORDER BY auth.name %s`, artIDs.StringList(), dbOrder)
	case "similar":
		q = fmt.Sprintf(
			`SELECT article_id
			FROM similar
			WHERE article_id IN (%s)
			GROUP BY (article_id)
			ORDER BY COUNT(*) %s`, artIDs.StringList(), dbOrder)
	default:
		return nil, fmt.Errorf("unsupported sort field '%s'", fieldName)
	}

	rows, err := store.db.Query(q)
	if err != nil {
		return nil, err
	}

	out := make(ArtList, 0, len(artIDs))
	for rows.Next() {
		var artID ArtID
		err = rows.Scan(&artID)
		if err != nil {
			return nil, err
		}
		out = append(out, artID)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return out, nil
}

//***************************
// support for sorting Articles

type By func(p1, p2 *Article) bool

func (by By) Sort(arts []*Article) {
	ps := &artSorter{
		arts: arts,
		by:   by,
	}
	sort.Sort(ps)
}

type artSorter struct {
	arts []*Article
	by   func(p1, p2 *Article) bool
}

// Len is part of sort.Interface.
func (s *artSorter) Len() int {
	return len(s.arts)
}

// Swap is part of sort.Interface.
func (s *artSorter) Swap(i, j int) {
	s.arts[i], s.arts[j] = s.arts[j], s.arts[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *artSorter) Less(i, j int) bool {
	return s.by(s.arts[i], s.arts[j])
}

type Iter struct {
	store *Store
	arts  ArtList
	pos   int
	cur   *Article
	err   error
}

func (it *Iter) Next() bool {
	if it.err != nil {
		return false
	}
	if it.pos >= len(it.arts) {
		return false // no more
	}

	var foo []*Article
	foo, it.err = it.store.Fetch(it.arts[it.pos])
	if it.err != nil {
		return false
	}
	it.cur = foo[0]
	it.pos++
	return true
}

func (it *Iter) Err() error {
	return it.err
}

func (it *Iter) Cur() *Article {
	return it.cur
}

func (store *Store) IterateAllArts() *Iter {
	all, err := store.AllArts()
	return &Iter{
		store: store,
		arts:  all,
		err:   err,
	}
}

// IterateArts returns an Iter which steps over each article in turn
func (store *Store) IterateArts(arts ...ArtID) *Iter {
	return &Iter{
		store: store,
		arts:  arts,
	}
}

func (store *Store) IterateTaggedArts() *Iter {
	it := &Iter{
		store: store,
		arts:  ArtList{},
	}

	rows, err := store.db.Query(`SELECT DISTINCT article_id FROM article_tag`)
	if err != nil {
		it.err = err
		return it
	}

	for rows.Next() {
		var artID ArtID
		err = rows.Scan(&artID)
		if err != nil {
			it.err = err
			rows.Close()
			break
		}
		it.arts = append(it.arts, artID)
	}
	err = rows.Err()
	if err != nil {
		it.err = err
	}
	return it
}

func (store *Store) Lang() string {
	return store.lang
}

// set the default indexing language for this store
// NOTE: won't take effect unless index is deleted and rebuilt...
func (store *Store) SetLang(lang string) error {

	err := validateLang(lang)
	if err != nil {
		return err
	}
	_, err = store.db.Exec(`UPDATE settings SET v=? WHERE n='lang'`, lang)
	if err != nil {
		return err
	}
	store.lang = lang

	// close, delete and rebuild index
	store.idx.Close()
	store.idx = nil
	indexDir := store.dbFile + ".bleve"
	store.dbug.Printf("Deleting %s\n", indexDir)
	err = os.RemoveAll(indexDir)
	if err != nil {
		return err
	}

	store.dbug.Printf("Recreate index\n")
	// new indexer
	store.idx, err = newBleveIndex(store.dbug, indexDir, store.lang, store.loc)
	if err != nil {
		return err
	}

	// index all articles
	allArts, err := store.AllArts()
	if err != nil {
		return err
	}
	err = store.reindex(allArts)
	if err != nil {
		return err
	}

	return nil
}
