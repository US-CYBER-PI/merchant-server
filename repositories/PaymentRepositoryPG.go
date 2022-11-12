package repositories

import (
	"MerchantServer/models"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type PaymentRepositoryPG struct {
	db        *sql.DB
	tableName string
}

func NewPaymentRepositoryPG(host, port, user, password, dbname, tableName string) (*PaymentRepositoryPG, error) {

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &PaymentRepositoryPG{
		db:        db,
		tableName: tableName,
	}, nil
}

func (p *PaymentRepositoryPG) Create(amount float64, UserID int, Status, payTokenIdentifier string) int {
	var id int
	err := p.db.QueryRow(fmt.Sprintf("INSERT INTO %s (amount, user_id, status, pay_token_identifier, bill_id) VALUES ($1, $2, $3, $4, '') RETURNING id", p.tableName), amount, UserID, Status, payTokenIdentifier).Scan(&id)
	if err != nil {
		panic(err)
		return -1
	}
	return id
}

func (p *PaymentRepositoryPG) UpdateStatus(id int, status string) bool {
	_, err := p.db.Exec(fmt.Sprintf("UPDATE %s SET status = $1 WHERE id = $2", p.tableName), status, id)
	if err != nil {
		return false
	}
	return true
}

func (p *PaymentRepositoryPG) UpdatePayment(id int, status, billId string) bool {
	_, err := p.db.Exec(fmt.Sprintf("UPDATE %s SET status = $1, bill_id = $2 WHERE id = $3", p.tableName), status, billId, id)
	if err != nil {
		return false
	}
	return true
}

func (p *PaymentRepositoryPG) GetById(id int) *models.Payment {
	var payment models.Payment
	err := p.db.QueryRow(fmt.Sprintf("SELECT id, amount, user_id, status, pay_token_identifier FROM %s WHERE id = $1", p.tableName), id).Scan(&payment.ID, &payment.Amount, &payment.UserID, &payment.Status, &payment.PayTokenIdentifier)
	if err != nil {
		return nil
	}
	return &payment
}
