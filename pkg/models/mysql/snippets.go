package mysql

import (
	"cosmasgithinji.net/simplesnippetbox/pkg/models"
	"database/sql"
)

type SnippetModel struct {
	DB *sql.DB //sql.DB connection pool
}

// Insert into db
func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	stmt := `INSERT INTO snippets (title, content, created, expires) 
			VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY));`
	result, err := m.DB.Exec(stmt, title, content, expires) //sql.Result object
	if err != nil {
		return 0, nil
	}
	id, err := result.LastInsertId() //latest ID
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Return snippet with id
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets 
			WHERE expires > UTC_TIMESTAMP() AND id = ?`
	row := m.DB.QueryRow(stmt, id) // hold a pointer to sql.ROW object
	s := &models.Snippet{}         //init a zeroed pointer to a Snippet struct

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}
	return s, nil // all OK, return snippet
}

// Return 10 most recent
func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets 
			WHERE expires > UTC_TIMESTAMP() ORDER BY created DESC LIMIT 10`
	rows, err := m.DB.Query(stmt) //return a sql.Rows resultset
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snippets := []*models.Snippet{} //empty slice to hold the models.Snippets objects

	for rows.Next() {
		s := &models.Snippet{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s) //append to snippet slice
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return snippets, nil // if everything is ok
}
