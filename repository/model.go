package repository

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Name     string `db:"name" json:"name"`
	Password []byte `db:"password" json:"-"`
}

func (u User) Validate(user, pass string) bool {
	return u.Name == user && bcrypt.CompareHashAndPassword(u.Password, []byte(pass)) == nil
}

func (u *User) SetPassword(pass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("cannot hash password: %w", err)
	}
	u.Password = hash
	return nil
}
