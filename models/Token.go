package models

type Token struct {
	Id          int
	Status      bool
	ExpiredDate string
	Token       string
}
