package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Connect creates and returns a PostgreSQL database connection using the provided configuration.
func Connect() (*sql.DB, error) {
	cfg := getConfig()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s %s",
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresPass,
		cfg.PostgresDefaultDb,
		cfg.PostgresUriArgs,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Optionally, you can ping the database to ensure the connection is established
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
