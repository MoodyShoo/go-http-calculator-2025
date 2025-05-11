package userrepo

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/MoodyShoo/go-http-calculator/internal/models"
)

type UserRepo struct {
	Db *sql.DB
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	return salt, err
}

func generatePasswordHash(password string, salt []byte) string {
	hash := sha256.Sum256(append([]byte(password), salt...))
	return hex.EncodeToString(hash[:])
}

func (ur *UserRepo) AddUser(login, password string) error {
	query := `INSERT INTO users (login, password, salt) VALUES ($1, $2, $3)`

	salt, err := generateSalt()
	if err != nil {
		return fmt.Errorf("generate salt: %w", err)
	}

	hashed := generatePasswordHash(password, salt)

	_, err = ur.Db.Exec(query, login, hashed, salt)
	if err != nil {
		return fmt.Errorf("user already exists")
	}

	return nil
}

func (ur *UserRepo) GetUser(login, password string) (models.User, error) {
	query := `SELECT * FROM users WHERE login = $1`

	var user models.User
	var dbHash string
	var salt []byte

	err := ur.Db.QueryRow(query, login).Scan(&user.Id, &user.Login, &dbHash, &salt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("user not found")
		}
		return models.User{}, fmt.Errorf("query error: %w", err)
	}

	hashedInput := generatePasswordHash(password, salt)

	if hashedInput != dbHash {
		return models.User{}, fmt.Errorf("invalid password")
	}

	return user, nil
}
