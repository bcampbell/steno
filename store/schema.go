package store

import (
	"database/sql"
)

// schemaVersion tries to retrieve the schema version from the database.
// It returns 0 if no schema at all (ie blank db).
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

// createSchema sets up the schema from scratch.
func (store *Store) createSchema() error {

	var err error
	// TODO:
	// full Publication
	_, err = store.db.Exec(`CREATE TABLE article (
        id INTEGER PRIMARY KEY,
        canonical_url TEXT NOT NULL,
        headline TEXT NOT NULL,
        content TEXT NOT NULL,
        published TEXT NOT NULL,
        updated TEXT NOT NULL,
        section TEXT NOT NULL DEFAULT '',
        retweet_count INTEGER,
        favourite_count INTEGER,
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
	_, err = store.db.Exec(`CREATE TABLE article_keyword (
         id INTEGER PRIMARY KEY,
         article_id INTEGER NOT NULL,   -- should be foreign key
         name TEXT NOT NULL,
         url TEXT NOT NULL )`)
	if err != nil {
		return err
	}

	_, err = store.db.Exec(`CREATE TABLE article_link (
         id INTEGER PRIMARY KEY,
         article_id INTEGER NOT NULL,   -- should be foreign key
         url TEXT NOT NULL )`)
	if err != nil {
		return err
	}

	_, err = store.db.Exec(`CREATE TABLE article_author (
         id INTEGER PRIMARY KEY,
         article_id INTEGER NOT NULL,   -- should be foreign key
         name TEXT NOT NULL )`)
	if err != nil {
		return err
	}

	// Note foreign key enforcement not on in sqlite by default, so
	// the ON DELETE CASCADE here is just documenting intention ;-)
	_, err = store.db.Exec(`CREATE TABLE similar (
         id INTEGER PRIMARY KEY,
		 article_id INTEGER NOT NULL,
		 other_id INTEGER NOT NULL,
		 score REAL NOT NULL DEFAULT 0.0,
		 FOREIGN KEY(article_id) REFERENCES article(id) ON DELETE CASCADE,
		 FOREIGN KEY(other_id) REFERENCES article(id) ON DELETE CASCADE )`)
	if err != nil {
		return err
	}

	// TODO: just do it all like this?
	indices := []string{
		`CREATE INDEX article_tag_artid ON article_tag(article_id)`,
		`CREATE INDEX article_url_artid ON article_url(article_id)`,
		`CREATE INDEX article_author_artid ON article_author(article_id)`,
		`CREATE INDEX article_keyword_artid ON article_keyword(article_id)`,
		`CREATE INDEX article_link_artid ON article_link(article_id)`,
	}
	for _, stmt := range indices {
		_, err = store.db.Exec(stmt)
		if err != nil {
			return err
		}

	}

	_, err = store.db.Exec(`CREATE TABLE settings (n TEXT, v TEXT NOT NULL)`)
	if err != nil {
		return err
	}
	_, err = store.db.Exec(`INSERT INTO settings (n,v) VALUES ('lang',?)`, store.lang)
	if err != nil {
		return err
	}

	_, err = store.db.Exec(`CREATE TABLE version (ver INTEGER NOT NULL)`)
	if err != nil {
		return err
	}

	// Set schema to current version
	_, err = store.db.Exec(`INSERT INTO version (ver) VALUES (7)`)
	if err != nil {
		return err
	}
	return nil
}

// ensureSchema makes sure the database schema is present and up-to-date.
func (store *Store) ensureSchema() error {
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
		store.dbug.Printf("updating database to version 2\n")
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

	if ver < 3 {
		store.dbug.Printf("updating database to version 3\n")
		_, err = store.db.Exec(`CREATE TABLE article_author (
            id INTEGER PRIMARY KEY,
            article_id INTEGER NOT NULL,   -- should be foreign key
            name TEXT NOT NULL)`)
		if err != nil {
			return err
		}
		_, err = store.db.Exec(`UPDATE version SET ver=3`)
		if err != nil {
			return err
		}
	}

	if ver < 4 {
		store.dbug.Printf("updating database to version 4\n")

		_, err = store.db.Exec(`ALTER TABLE article ADD COLUMN retweet_count INTEGER NOT NULL DEFAULT 0`)
		if err != nil {
			return err
		}

		_, err = store.db.Exec(`ALTER TABLE article ADD COLUMN favourite_count INTEGER NOT NULL DEFAULT 0`)
		if err != nil {
			return err
		}

		_, err = store.db.Exec(`CREATE TABLE article_keyword (
         id INTEGER PRIMARY KEY,
         article_id INTEGER NOT NULL,   -- should be foreign key
         name TEXT NOT NULL,
         url TEXT NOT NULL )`)
		if err != nil {
			return err
		}
		_, err = store.db.Exec(`CREATE TABLE article_link (
         id INTEGER PRIMARY KEY,
         article_id INTEGER NOT NULL,   -- should be foreign key
         url TEXT NOT NULL )`)
		if err != nil {
			return err
		}

		_, err = store.db.Exec(`UPDATE version SET ver=4`)
		if err != nil {
			return err
		}
	}

	if ver < 5 {
		store.dbug.Printf("updating database to version 5\n")
		indices := []string{
			`CREATE INDEX article_tag_artid ON article_tag(article_id)`,
			`CREATE INDEX article_url_artid ON article_url(article_id)`,
			`CREATE INDEX article_author_artid ON article_author(article_id)`,
			`CREATE INDEX article_keyword_artid ON article_keyword(article_id)`,
			`CREATE INDEX article_link_artid ON article_link(article_id)`,
		}
		for _, stmt := range indices {
			_, err = store.db.Exec(stmt)
			if err != nil {
				return err
			}

		}

		_, err = store.db.Exec(`UPDATE version SET ver=5`)
		if err != nil {
			return err
		}
	}

	if ver < 6 {
		store.dbug.Printf("updating database to version 6\n")
		_, err = store.db.Exec(`CREATE TABLE settings (n TEXT, v TEXT NOT NULL)`)
		if err != nil {
			return err
		}
		_, err = store.db.Exec(`INSERT INTO settings (n,v) VALUES ('lang',?)`, store.lang)
		if err != nil {
			return err
		}

		_, err = store.db.Exec(`UPDATE version SET ver=6`)
		if err != nil {
			return err
		}
	}

	if ver < 7 {
		store.dbug.Printf("updating database to version 7\n")
		_, err = store.db.Exec(`CREATE TABLE similar (
         id INTEGER PRIMARY KEY,
		 article_id INTEGER NOT NULL,
		 other_id INTEGER NOT NULL,
		 score REAL NOT NULL DEFAULT 0.0,
		 FOREIGN KEY(article_id) REFERENCES article(id) ON DELETE CASCADE,
		 FOREIGN KEY(other_id) REFERENCES article(id) ON DELETE CASCADE )`)
		if err != nil {
			return err
		}
		_, err = store.db.Exec(`UPDATE version SET ver=7`)
		if err != nil {
			return err
		}
	}
	return nil
}
