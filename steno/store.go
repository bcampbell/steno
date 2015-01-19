package main

import (
	"database/sql"
	"fmt"
	//"github.com/bcampbell/arts/arts"
	"github.com/bcampbell/badger"
	"github.com/bcampbell/badger/query"
	_ "github.com/mattn/go-sqlite3"
	"regexp"
	"strings"
)

//
var defaultField string = "content"

//
type Store struct {
	db   *sql.DB
	coll *badger.Collection
}

func NewStore(dbFile string) (*Store, error) {
	store := &Store{}

	var err error
	store.db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	err = store.initDB()
	if err != nil {
		return nil, err
	}

	arts, err := store.readAllArts()
	if err != nil {
		return nil, err
	}
	store.coll = badger.NewCollection(&Article{})
	for _, art := range arts {
		store.coll.Put(art)
	}
	return store, nil
}

func DummyStore() *Store {
	store := &Store{}
	store.coll = badger.NewCollection(&Article{})
	return store
}

func (store *Store) Close() {
	if store.db != nil {
		dbug.Printf("Close sqlite db\n")
		store.db.Close()
		store.db = nil
	}
	store.coll = badger.NewCollection(&Article{})
}

func (store *Store) TotalArts() int {
	return store.coll.Count()
}

func (store *Store) schemaVersion() (int, error) {
	var n string
	err := store.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='article';`).Scan(&n)
	if err == sql.ErrNoRows {
		return 0, nil // no schema at all
	}
	err = store.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='version';`).Scan(&n)
	if err == sql.ErrNoRows {
		return 1, nil // version 1: no version table :-)
	}
	if err != nil {
		return 0, err
	}

	var v int
	err = store.db.QueryRow(`SELECT MAX(ver) FROM version`).Scan(&v)
	if err != nil {
		return 0, err
	}

	return v, nil
}

func (store *Store) createSchema() error {

	var err error
	// TODO:
	// Authors
	// full Publication
	// Keywords
	_, err = store.db.Exec(`CREATE TABLE article (
         id INTEGER PRIMARY KEY,
         canonical_url TEXT NOT NULL,
         headline TEXT NOT NULL,
         content TEXT NOT NULL,
         published TEXT NOT NULL,
         updated TEXT NOT NULL,
         section TEXT NOT NULL DEFAULT '',
         pub TEXT NOT NULL )`)
	if err != nil {
		return err
	}

	_, err = store.db.Exec(`CREATE TABLE article_tag (
         id INTEGER PRIMARY KEY,
         article_id INTEGER NOT NULL,   -- should be foreign key
         tag TEXT NOT NULL )`)
	if err != nil {
		return err
	}
	_, err = store.db.Exec(`CREATE TABLE article_url (
         id INTEGER PRIMARY KEY,
         article_id INTEGER NOT NULL,   -- should be foreign key
         url TEXT NOT NULL )`)
	if err != nil {
		return err
	}

	_, err = store.db.Exec(`CREATE TABLE version (ver INTEGER NOT NULL)`)
	if err != nil {
		return err
	}
	_, err = store.db.Exec(`INSERT INTO version (ver) VALUES (2)`)
	if err != nil {
		return err
	}

	return nil
}

func (store *Store) initDB() error {
	ver, err := store.schemaVersion()
	if err != nil {
		return err
	}

	if ver == 0 {
		// blank slate
		return store.createSchema()
	}

	// schema exists - apply any migrations required
	if ver < 2 {
		dbug.Printf("updating database to version 2\n")
		_, err = store.db.Exec(`ALTER TABLE article ADD COLUMN section TEXT NOT NULL DEFAULT ''`)
		if err != nil {
			return err
		}
		_, err = store.db.Exec(`CREATE TABLE version (ver INTEGER NOT NULL)`)
		if err != nil {
			return err
		}
		_, err = store.db.Exec(`INSERT INTO version (ver) VALUES (2)`)
		if err != nil {
			return err
		}
	}
	return nil
}

// read in articles from DB
func (store *Store) readAllArts() (ArtList, error) {
	db := store.db

	tab := map[int]*Article{}

	rows, err := db.Query("SELECT id,canonical_url,headline,content,published,updated,pub,section FROM article")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		a := &Article{}
		err = rows.Scan(&a.ID, &a.CanonicalURL, &a.Headline, &a.Content, &a.Published, &a.Updated, &a.Pub, &a.Section)
		if err != nil {
			return nil, err
		}
		tab[a.ID] = a
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	//
	rows, err = db.Query("SELECT article_id,url FROM article_url")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID int
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

	//
	rows, err = db.Query("SELECT article_id,tag FROM article_tag")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var artID int
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

	out := ArtList{}
	for _, a := range tab {
		out = append(out, a)
	}
	return out, nil
}

//standin - return all articles
func (store *Store) AllArts() (ArtList, error) {
	q := badger.NewAllQuery()
	var arts ArtList
	store.coll.Find(q, &arts)

	publishedDesc := func(a1, a2 *Article) bool {
		return a1.Published > a2.Published
	}
	By(publishedDesc).Sort(arts)

	return arts, nil
}

// search performs a search and returns the results
func (store *Store) Search(queryString string) (ArtList, error) {
	q, err := query.Parse(queryString, store.coll.ValidFields(), defaultField)
	if err != nil {
		return nil, err
	}
	// TODO: fix badger so it's not so silly!
	if q == nil {
		return store.AllArts()
	}

	var arts ArtList
	store.coll.Find(q, &arts)

	publishedDesc := func(a1, a2 *Article) bool {
		return a1.Published > a2.Published
	}
	By(publishedDesc).Sort(arts)

	return arts, nil
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

func (store *Store) AddTag(arts ArtList, tag string) (ArtList, error) {
	tag = strings.ToLower(tag)

	if store.db == nil {
		return ArtList{}, nil
	}

	// apply to db
	tx, err := store.db.Begin()
	if err != nil {
		return nil, err
	}

	delStmt, err := tx.Prepare("DELETE FROM article_tag WHERE article_id=? AND tag=?")
	if err != nil {
		return nil, err
	}
	defer delStmt.Close()
	insStmt, err := tx.Prepare("INSERT INTO article_tag(article_id,tag) VALUES(?,?)")
	if err != nil {
		return nil, err
	}
	defer insStmt.Close()

	for _, art := range arts {
		_, err = delStmt.Exec(art.ID, tag)
		if err != nil {
			return nil, err
		}
		_, err = insStmt.Exec(art.ID, tag)
		if err != nil {
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// apply to index
	affected := ArtList{}
	for _, art := range arts {
		if art.AddTag(tag) {
			affected = append(affected, art)
		}
	}

	return affected, nil
}

func (store *Store) RemoveTag(arts ArtList, tag string) (ArtList, error) {
	tag = strings.ToLower(tag)

	if store.db == nil {
		return ArtList{}, nil
	}

	// apply to db
	tx, err := store.db.Begin()
	if err != nil {
		return nil, err
	}

	delStmt, err := tx.Prepare("DELETE FROM article_tag WHERE article_id=? AND tag=?")
	if err != nil {
		return nil, err
	}
	defer delStmt.Close()

	for _, art := range arts {
		_, err = delStmt.Exec(art.ID, tag)
		if err != nil {
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// apply to index
	affected := ArtList{}
	for _, art := range arts {
		if art.RemoveTag(tag) {
			affected = append(affected, art)
		}
	}

	return affected, nil
}

// delete articles
func (store *Store) Delete(arts ArtList) error {
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}

	affected, err := store.doDelete(tx, arts)
	if err != nil {
		dbug.Printf("error, rolling back\n")
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	dbug.Printf("Deleted %d articles\n", affected)
	return nil
}

func (store *Store) doDelete(tx *sql.Tx, arts ArtList) (int64, error) {
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
	delArtStmt, err := tx.Prepare("DELETE FROM article WHERE id=?")
	if err != nil {
		return 0, err
	}
	defer delArtStmt.Close()

	for _, art := range arts {

		_, err = delTagsStmt.Exec(art.ID)
		if err != nil {
			return 0, err
		}

		_, err = delURLsStmt.Exec(art.ID)
		if err != nil {
			return 0, err
		}
		r, err := delArtStmt.Exec(art.ID)
		if err != nil {
			return 0, err
		}
		// check we actually deleted something...
		foo, err := r.RowsAffected()
		if err != nil {
			return 0, err
		}
		affected += foo
	}

	// now update the index
	for _, art := range arts {
		store.coll.Remove(art)
	}
	return affected, nil
}

/*
func getPublications() ([]string, error) {
	var arts []*Article
	coll.Find(badger.NewAllQuery(), &arts)
	pubSet := make(map[string]struct{})
	for _, art := range arts {
		pubSet[art.Pub] = struct{}{}
	}
	var pubs []string
	for pub, _ := range pubSet {
		if pub != "" {
			pubs = append(pubs, pub)
		}
	}
	sort.Strings(pubs)

	return pubs, nil
}
*/

func (store *Store) FindArt(urls []string) (int, error) {
	placeholders := make([]string, len(urls))
	params := make([]interface{}, len(urls))
	for i, _ := range urls {
		placeholders[i] = "?"
		params[i] = urls[i]
	}
	foo := fmt.Sprintf(`SELECT article_id FROM article_url WHERE url IN(%s)`, strings.Join(placeholders, ","))

	var artID int
	err := store.db.QueryRow(foo, params...).Scan(&artID)
	if err == sql.ErrNoRows {
		return 0, nil // article not found
	} else if err != nil {
		return 0, err
	} else {
		return artID, nil
	}
}

func (store *Store) Stash(art *Article) error {
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}

	if art.ID != 0 {
		panic("article id already set")
	}
	err = store.doStash(tx, art)
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
	result, err := tx.Exec(`INSERT INTO article(canonical_url, headline, content,published,updated,pub,section)
        values(?,?,?,?,?,?,?)`,
		art.CanonicalURL,
		art.Headline,
		art.Content,
		art.Published,
		art.Updated,
		art.Pub,
		art.Section)
	if err != nil {
		return err
	}

	if artID, err := result.LastInsertId(); err == nil {
		art.ID = int(artID)
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
	// TODO: authors, keywords

	store.coll.Put(art)

	return nil
}
