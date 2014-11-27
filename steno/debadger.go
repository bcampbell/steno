package main

import (
	"database/sql"
	"fmt"
	//"github.com/bcampbell/arts/arts"
	//	"github.com/bcampbell/badger"
	_ "github.com/mattn/go-sqlite3"
)

// write articles out to db
func debadger(srcArts ArtList, dbFile string) error {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	// TODO:
	// Authors
	// Publication
	// Keywords
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS article (
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS article_tag (
         id INTEGER PRIMARY KEY,
         article_id INTEGER NOT NULL,   -- should be foreign key
         tag TEXT NOT NULL )`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS article_url (
         id INTEGER PRIMARY KEY,
         article_id INTEGER NOT NULL,   -- should be foreign key
         url TEXT NOT NULL )`)
	if err != nil {
		return err
	}

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

// read in files from DB
func enbadger(dbFile string) (ArtList, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tab := map[int]*Article{}

	rows, err := db.Query("SELECT id,canonical_url,headline,content,published,updated,pub FROM article")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		a := &Article{}
		var id int
		err = rows.Scan(&id, &a.CanonicalURL, &a.Headline, &a.Content, &a.Published, &a.Updated, &a.Pub)
		if err != nil {
			return nil, err
		}
		tab[id] = a
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	//
	fmt.Printf(" reading urls...\n")

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

	fmt.Printf(" reading tags...\n")
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

	fmt.Printf(" done.\n")

	out := ArtList{}
	for _, a := range tab {
		out = append(out, a)
	}
	return out, nil
}
