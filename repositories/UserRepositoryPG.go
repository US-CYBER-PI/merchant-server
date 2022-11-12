package repositories

import (
	"MerchantServer/models"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type UserRepositoryPG struct {
	db             *sql.DB
	tableUserName  string
	tableTokenName string
}

func NewUserRepositoryPG(host, port, user, password, dbname, tableUserName, tableTokenName string) (*UserRepositoryPG, error) {

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &UserRepositoryPG{
		db:             db,
		tableUserName:  tableUserName,
		tableTokenName: tableTokenName,
	}, nil
}

func (u *UserRepositoryPG) GetUserById(id int) *models.User {
	var user models.User
	err := u.db.QueryRow(fmt.Sprintf("SELECT id, phone, token_id FROM %s WHERE id = $1", u.tableUserName), id).Scan(&user.Id, &user.Phone, &user.TokenId)
	if err != nil {
		return nil
	}
	return &user
}

func (u *UserRepositoryPG) GetTokenById(id int) *models.Token {
	var token models.Token
	err := u.db.QueryRow(fmt.Sprintf("SELECT id, token FROM %s WHERE id = $1", u.tableTokenName), id).Scan(&token.Id, &token.Token)
	if err != nil {
		return nil
	}
	return &token
}
