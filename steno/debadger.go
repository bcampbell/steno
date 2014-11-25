package main

import (
	"database/sql"
	"fmt"
	//"github.com/bcampbell/arts/arts"
	//	"github.com/bcampbell/badger"
	_ "github.com/mattn/go-sqlite3"
)

func debadger(srcArts ArtList, outFile string) error {
	db, err := sql.Open("sqlite3", outFile)
	if err != nil {
		return err
	}
	defer db.Close()

	// TODO:
	// Authors
	// URLs
	// Publication
	// Keywords
	// Tags
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

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into article(canonical_url, headline, content,published,updated,pub) values(?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, art := range srcArts {
		_, err = stmt.Exec(art.CanonicalURL, art.Headline, art.Content, art.Published, art.Updated, art.Pub)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func enbadger(inFile string) (ArtList, error) {
	db, err := sql.Open("sqlite3", inFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id,canonical_url,headline,content,published,updated,pub FROM article")
	if err != nil {
		return nil, err
	}
	out := ArtList{}
	for rows.Next() {
		a := &Article{}
		var id int
		err = rows.Scan(&id, &a.CanonicalURL, &a.Headline, &a.Content, &a.Published, &a.Updated, &a.Pub)
		if err != nil {
			return nil, err
		}
		fmt.Println(id, a.Headline)
		out = append(out, a)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return out, nil
}
