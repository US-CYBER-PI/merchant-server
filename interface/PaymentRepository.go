package _interface

import "MerchantServer/models"

type PaymentRepository interface {
	Create(amount float64, UserID int, Status, payTokenIdentifier string) int
	UpdateStatus(id int, status string) bool
	UpdatePayment(id int, status, billId string) bool
	GetById(id int) *models.Payment
}
