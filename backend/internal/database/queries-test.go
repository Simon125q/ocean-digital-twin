package database

import (
	"fmt"
	"log/slog"
)

func (s *service) GetCount() int {
	query := `
    SELECT count
    FROM test
    ORDER BY id DESC
    LIMIT 1;
    `
	var count int
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		slog.Error("Error quering test for count")
		return 0
	}
	return count
}

func (s *service) UpdateCount(newCount int) error {
	query := `
    UPDATE test
    SET count = $1
    WHERE id = $2
    `
	_, err := s.db.Exec(query, newCount, 1)
	if err != nil {
		return fmt.Errorf("error updating count: %w", err)
	}
	return nil
}

func (s *service) NewCount() (int, error) {
	query := `
    INSERT INTO test (count)
    VALUES ($1)
    RETURNING id
    `
	var id int
	err := s.db.QueryRow(query, 0).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
