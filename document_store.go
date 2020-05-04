package tinysearch

import "database/sql"

type DocumentStore struct {
	db *sql.DB
}

func NewDocumentStore(db *sql.DB) *DocumentStore {
	return &DocumentStore{db: db}
}

func (ds *DocumentStore) save(title string) (DocumentID, error) {
	query := "INSERT INTO documents (document_title) VALUES (?)"
	result, err := ds.db.Exec(query, title)
	id, err := result.LastInsertId()
	return DocumentID(id), err
}
