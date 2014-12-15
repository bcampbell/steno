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

	err = store.initDBSchema()
	if err != nil {
		return nil, err
	}

	arts, err := store.readAllArts()
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
		store.db.Close()
		store.db = nil
	}
	store.coll = badger.NewCollection(&Article{})
}

func (store *Store) TotalArts() int {
	return store.coll.Count()
}

func (store *Store) initDBSchema() error {

	var err error
	// TODO: store a schema version number in db

	// TODO:
	// Authors
	// Publication
	// Keywords
	_, err = store.db.Exec(`CREATE TABLE IF NOT EXISTS article (
         id INTEGER PRIMARY KEY,
         canonical_url TEXT NOT NULL,
         headline TEXT NOT NULL,
         content TEXT NOT NULL,
         published TEXT NOT NULL,
         updated TEXT NOT NULL,
         pub TEXT NOT NULL )`)
	if err != nil {
		return err
	}

	_, err = store.db.Exec(`CREATE TABLE IF NOT EXISTS article_tag (
         id INTEGER PRIMARY KEY,
         article_id INTEGER NOT NULL,   -- should be foreign key
         tag TEXT NOT NULL )`)
	if err != nil {
		return err
	}
	_, err = store.db.Exec(`CREATE TABLE IF NOT EXISTS article_url (
         id INTEGER PRIMARY KEY,
         article_id INTEGER NOT NULL,   -- should be foreign key
         url TEXT NOT NULL )`)
	if err != nil {
		return err
	}

	return nil
}

// write articles out to db
/*
func debadger(srcArts ArtList, dbFile string) error {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	artStmt, err := tx.Prepare("insert into article(canonical_url, headline, content,published,updated,pub) values(?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer artStmt.Close()

	urlStmt, err := tx.Prepare("insert into article_url(article_id,url) values(?,?)")
	if err != nil {
		return err
	}
	defer urlStmt.Close()

	tagStmt, err := tx.Prepare("insert into article_tag(article_id,tag) values(?,?)")
	if err != nil {
		return err
	}
	defer tagStmt.Close()

	for _, art := range srcArts {
		var result sql.Result
		result, err = artStmt.Exec(art.CanonicalURL, art.Headline, art.Content, art.Published, art.Updated, art.Pub)
		if err != nil {
			return err
		}
		artID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		for _, u := range art.URLs {
			_, err = urlStmt.Exec(artID, u)
			if err != nil {
				return err
			}
		}
		for _, tag := range art.Tags {
			_, err = tagStmt.Exec(artID, tag)
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
}
*/

// read in files from DB
func (store *Store) readAllArts() (ArtList, error) {
	db := store.db

	tab := map[int]*Article{}

	rows, err := db.Query("SELECT id,canonical_url,headline,content,published,updated,pub FROM article")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		a := &Article{}
		err = rows.Scan(&a.ID, &a.CanonicalURL, &a.Headline, &a.Content, &a.Published, &a.Updated, &a.Pub)
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
	result, err := tx.Exec(`INSERT INTO article(canonical_url, headline, content,published,updated,pub)
        values(?,?,?,?,?,?)`,
		art.CanonicalURL,
		art.Headline,
		art.Content,
		art.Published,
		art.Updated,
		art.Pub)
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
