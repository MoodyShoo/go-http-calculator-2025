package database

import (
	"database/sql"

	expressionrepo "github.com/MoodyShoo/go-http-calculator/internal/database/repository/expression_repo"
	userrepo "github.com/MoodyShoo/go-http-calculator/internal/database/repository/user_repo"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db             *sql.DB
	ExpressionRepo *expressionrepo.ExpressionRepo
	UserRepo       *userrepo.UserRepo
}

func (d *Database) createTables() error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		login TEXT UNIQUE NOT NULL,
		password BLOB NOT NULL,
		salt BLOB NOT NULL
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		expression TEXT NOT NULL,
		status TEXT NOT NULL,
		result REAL,
		error TEXT,
		user_id INTEGER NOT NULL,
	
		FOREIGN KEY (user_id)  REFERENCES  users (id)
	);`
	)

	if _, err := d.db.Exec(usersTable); err != nil {
		return err
	}

	if _, err := d.db.Exec(expressionsTable); err != nil {
		return err
	}

	return nil
}

func NewInMemoryDatabase() (*Database, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	database := &Database{
		db: db,
		ExpressionRepo: &expressionrepo.ExpressionRepo{
			Db: db,
		},
		UserRepo: &userrepo.UserRepo{
			Db: db,
		},
	}

	if err = database.createTables(); err != nil {
		return nil, err
	}

	return database, nil
}

func NewDatabase() (*Database, error) {
	db, err := sql.Open("sqlite3", "calculator.db")
	if err != nil {
		return nil, err
	}

	database := &Database{
		db: db,
		ExpressionRepo: &expressionrepo.ExpressionRepo{
			Db: db,
		},
		UserRepo: &userrepo.UserRepo{
			Db: db,
		},
	}

	if err = database.createTables(); err != nil {
		return nil, err
	}

	return database, nil
}
