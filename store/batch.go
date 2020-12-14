package store

import (
	"database/sql"
)

// Batch is a helper for performing operations upon the store.
// TODO:
// - add logging for bad/partly bad articles
// - track articles we already had
type Batch struct {
	store *Store
	//arts      []*Article
	// StashedIDs contains IDs of successfully-stashed articles
	//StashedIDs ArtList
	tx *sql.Tx
}

func (store *Store) BeginBatch() (*Batch, error) {

	tx, err := store.db.Begin()
	if err != nil {
		return nil, err
	}

	batch := &Batch{
		store: store,
		tx:    tx,
	}

	return batch, err
}

func (batch *Batch) ClearSimilar() error {
	_, err := batch.tx.Exec(`DELETE FROM similar`)
	if err != nil {
		return err
	}
	return nil
}

func (batch *Batch) AddSimilar(id ArtID, matches []Match) error {
	insStmt, err := batch.tx.Prepare("INSERT INTO similar(article_id, other_id, score) VALUES(?,?,?)")
	if err != nil {
		return err
	}
	defer insStmt.Close()

	for _, match := range matches {
		_, err = insStmt.Exec(id, match.ID, match.Score)
		if err != nil {
			return err
		}
	}
	return nil
}

func (batch *Batch) Commit() error {
	return batch.tx.Commit()
}

func (batch *Batch) Rollback() error {
	return batch.tx.Rollback()
}
