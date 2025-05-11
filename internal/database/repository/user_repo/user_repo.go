package userrepo

import "database/sql"

type UserRepo struct {
	Db *sql.DB
}
