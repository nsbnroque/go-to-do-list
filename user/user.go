package user

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/validator.v2"
)

type User struct {
	Name     string `json:"name" validate:"nonzero"`
	Email    string `json:"email" validate:"nonzero"`
	Password string `json:"password" validate:"nonzero,min=8"`
	Score    int64  `json:"score"`
}

func ValidateUser(user *User) error {
	if err := validator.Validate(user); err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
