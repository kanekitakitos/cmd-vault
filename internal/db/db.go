package db

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/kanekitakitos/cmd-vault/internal/models"
)

const schema = `
CREATE TABLE IF NOT EXISTS commands (
	id INTEGER PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	command_str TEXT NOT NULL,
	note TEXT NOT NULL,
	usage_count INTEGER DEFAULT 0,
	created_at TEXT NOT NULL
);
`

type Store struct {
	conn *sql.DB
}

func Open(path string) (*Store, error) {
	conn, err := sql.Open("sqlite3", path+"?_busy_timeout=5000&_foreign_keys=ON")
	if err != nil {
		return nil, err
	}
	s := &Store{conn: conn}
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	if s.conn == nil {
		return nil
	}
	return s.conn.Close()
}

func (s *Store) InsertCommand(c *models.Command) (int64, error) {
	if c.Name == "" {
		return 0, errors.New("name is required")
	}
	if c.Note == "" {
		return 0, errors.New("note is required")
	}
	stmt := `INSERT INTO commands (name, command_str, note, usage_count, created_at) VALUES (?, ?, ?, ?, ?)`
	res, err := s.conn.Exec(stmt, c.Name, c.CommandStr, c.Note, c.UsageCount, c.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetAllCommands() ([]models.Command, error) {
	rows, err := s.conn.Query(`SELECT id, name, command_str, note, usage_count, created_at FROM commands ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Command
	for rows.Next() {
		c, err := scanCommand(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (s *Store) GetByName(name string) (*models.Command, error) {
	row := s.conn.QueryRow(`SELECT id, name, command_str, note, usage_count, created_at FROM commands WHERE name = ?`, name)
	c, err := scanCommand(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// scanCommand is a helper to scan a command from a sql.Row or sql.Rows.
func scanCommand(s interface{ Scan(...interface{}) error }) (models.Command, error) {
	var c models.Command
	var createdAt string
	if err := s.Scan(&c.ID, &c.Name, &c.CommandStr, &c.Note, &c.UsageCount, &createdAt); err != nil {
		return models.Command{}, err
	}
	parsedTime, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return models.Command{}, err
	}
	c.CreatedAt = parsedTime
	return c, nil
}

func (s *Store) UpdateCommand(c *models.Command) error {
	if c == nil {
		return errors.New("nil command")
	}
	_, err := s.conn.Exec(`UPDATE commands SET name=?, command_str=?, note=?, usage_count=? WHERE id=?`, c.Name, c.CommandStr, c.Note, c.UsageCount, c.ID)
	return err
}

func (s *Store) DeleteCommand(id int) error {
	_, err := s.conn.Exec(`DELETE FROM commands WHERE id=?`, id)
	return err
}

func (s *Store) IncrementUsage(id int) error {
	_, err := s.conn.Exec(`UPDATE commands SET usage_count = usage_count + 1 WHERE id = ?`, id)
	return err
}
