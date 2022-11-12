package _interface

import "MerchantServer/models"

type UserRepository interface {
	GetUserById(id int) *models.User

	GetTokenById(id int) *models.Token
}
