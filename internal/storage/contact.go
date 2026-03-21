package storage

import (
	"context"
	"time"
)

const (
	ContactField string = "contact"
)

type Contact struct {
	ID     uint
	UserID uint

	Name     string
	Birthday time.Time

	Note string

	Phones []Phone
}

// todo: remove InitialPhones as PhoneStorage will be implemented
type AddContactData struct {
	//ID will be overwritten, Phones will be ignored
	Contact Contact
	//ID, ContactID will be overwritten
	InitialPhones    []Phone
	PhoneConstraints PhoneConstraints
}

type ContactPhonesPreload struct {
	Enabled     bool
	PrimaryOnly bool
}

// todo: remove Preload as PhoneStorage will be implemented
type GetContactData struct {
	ID     uint
	UserID uint

	Preload  ContactPhonesPreload
	WithNote bool
}

type GetContactsData struct {
	Selector Selector
	//Data.ID will be ignored
	Data GetContactData
}

type DeleteContactData struct {
	ID     uint
	UserID uint
}

type UpdateContactData struct {
	//Empty fields will be ignored, Phones field will be ignored
	Contact Contact
}

type ContactStorage interface {
	AddContact(ctx context.Context, data AddContactData) (*Contact, error)
	GetContact(ctx context.Context, data GetContactData) (*Contact, error)
	GetContacts(ctx context.Context, data GetContactsData) ([]Contact, error)
	UpdateContact(ctx context.Context, data UpdateContactData) error
	DeleteContact(ctx context.Context, data DeleteContactData) error
}
