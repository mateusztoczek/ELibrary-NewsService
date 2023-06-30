package database

import (
	"database/sql"
	"fmt"
	"news/config"

	"github.com/pkg/errors"
)

func ConnectDB(config config.Config) (*sql.DB, error) {

	dbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", config.Host, config.Port, config.User, config.Password, config.DBName)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to the database")
	}
	return db, nil
}

func CreateNewsTable(db *sql.DB, config config.Config) error {
	query := fmt.Sprintf(`
		CREATE SCHEMA IF NOT EXISTS "%s";
		CREATE TABLE IF NOT EXISTS "%s"."%s" (
			"Id" SERIAL PRIMARY KEY,
			"Content" TEXT NOT NULL,
			"CreatedDate" TIMESTAMP NOT NULL,
			"LastUpdate" TIMESTAMP NOT NULL,
			"AuthorId" TEXT NOT NULL
		);`, config.SchemaName, config.SchemaName, config.TableName)

	_, err := db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "failed to create News table")
	}
	return nil
}
