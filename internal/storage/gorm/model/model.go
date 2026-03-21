package model

import (
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/storage"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Login        string `gorm:"uniqueIndex"`
	PasswordAlgo string
	PasswordHash string

	Contacts []Contact `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (u *User) FromUser(user storage.User) {
	u.Login = user.Login
	u.PasswordAlgo = user.Password.Algo
	u.PasswordHash = user.Password.Hash
}

func (u User) ToUser() *storage.User {
	return &storage.User{
		ID:    u.ID,
		Login: u.Login,
		Password: storage.UserPassword{
			Hash: u.PasswordHash,
			Algo: u.PasswordAlgo,
		},
	}
}

type Contact struct {
	gorm.Model

	UserID   uint `gorm:"index"`
	Name     string
	Birthday time.Time
	Note     string

	Phones []Phone `gorm:"foreignKey:ContactID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (c *Contact) FromContact(contact storage.Contact) {
	c.Name = contact.Name
	c.Birthday = contact.Birthday
	c.Note = contact.Note
}

func (c Contact) ToContact() *storage.Contact {
	return &storage.Contact{
		ID:       c.ID,
		UserID:   c.UserID,
		Name:     c.Name,
		Birthday: c.Birthday,
		Note:     c.Note,
	}
}

type Phone struct {
	gorm.Model

	ContactID uint `gorm:"index"`
	Phone     string
	Primary   bool `gorm:"index"`
}

func (p *Phone) FromPhone(phone storage.Phone) {
	p.Phone = phone.Phone
	p.Primary = phone.Primary
}

func (p Phone) ToPhone() *storage.Phone {
	return &storage.Phone{
		ID:        p.ID,
		ContactID: p.ContactID,
		Phone:     p.Phone,
		Primary:   p.Primary,
	}
}
