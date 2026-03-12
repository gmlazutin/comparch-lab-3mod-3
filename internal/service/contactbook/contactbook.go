package contactbook

import (
	"context"
	"fmt"
	"time"

	"github.com/gmlazutin/comparch-lab-3mod-3/internal/logging"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/storage"

	"github.com/nyaruka/phonenumbers"
)

type Selector struct {
	Limit  uint
	Offset uint
}

func (s Selector) validate() error {
	if s.Limit == 0 {
		return service.ErrIncorrectSelectorValues
	}
	return nil
}

type ContactInfo struct {
	Name     string
	Birthday time.Time
	Note     string
}

func (ci ContactInfo) validate(ts time.Time) error {
	if ci.Birthday.After(ts) {
		return service.ErrIncorrectBirthday
	}

	return nil
}

type ContactID struct {
	ID     uint
	UserID uint
}

type Contact struct {
	ContactID ContactID
	Info      ContactInfo
}

type ContactWithPhones struct {
	Contact Contact
	Phones  []Phone
}

func (p *Contact) fromStorage(contact storage.Contact) {
	p.ContactID.ID = contact.ID
	p.ContactID.UserID = contact.UserID
	p.Info.Birthday = contact.Birthday
	p.Info.Name = contact.Name
	p.Info.Note = contact.Note
}

type PhoneInfo struct {
	Phone   string
	Primary bool
}

func (pi PhoneInfo) validate() error {
	p, err := phonenumbers.Parse(pi.Phone, "")
	if err != nil {
		return service.ErrIncorrectPhone
	}
	if !phonenumbers.IsValidNumber(p) {
		return service.ErrIncorrectPhone
	}

	return nil
}

type PhoneID struct {
	ID        uint
	ContactID uint
}

type Phone struct {
	PhoneID PhoneID
	Info    PhoneInfo
}

func (p *Phone) fromStorage(phone storage.Phone) {
	p.PhoneID.ID = phone.ID
	p.PhoneID.ContactID = phone.ContactID
	p.Info.Phone = phone.Phone
	p.Info.Primary = phone.Primary
}

type PhonesPreload struct {
	PrimaryOnly bool
}

type Options struct {
	ContactStorage storage.ContactStorage
	PhoneStorage   storage.UserStorage
	ServiceOpts    service.ServiceOptions
}

type Service struct {
	opts Options
}

var timeNow = time.Now

func New(options Options) *Service {
	if options.ServiceOpts.Logger == nil {
		options.ServiceOpts.Logger = logging.EmptyLogger()
	}
	return &Service{
		opts: options,
	}
}

func (s *Service) wrapErr(err error) error {
	return fmt.Errorf("contactbookService: %w", err)
}

func (s *Service) AddContact(ctx context.Context, uid uint, contact ContactInfo, phones []PhoneInfo) (*Contact, []Phone, error) {
	if err := contact.validate(timeNow()); err != nil {
		return nil, nil, s.wrapErr(fmt.Errorf("unable to validate contact: %w", err))
	}
	repophones := make([]storage.Phone, len(phones))
	for i := range phones {
		if err := phones[i].validate(); err != nil {
			return nil, nil, s.wrapErr(fmt.Errorf("unable to validate phone: %w", err))
		}
		repophones[i] = storage.Phone{
			Phone:   phones[i].Phone,
			Primary: phones[i].Primary,
		}
	}

	cont, err := s.opts.ContactStorage.AddContact(ctx, storage.AddContactData{
		Contact: storage.Contact{
			UserID:   uid,
			Name:     contact.Name,
			Birthday: contact.Birthday,
			Note:     contact.Note,
		},
		InitialPhones: repophones,
		PhoneConstraints: storage.PhoneConstraints{
			MaxAllowed: 10,
			MaxPrimaries: 1,
			MinPrimaries: 1,
		},
	})
	if err != nil {
		err = service.TranslateStorageError(err)
		return nil, nil, s.wrapErr(fmt.Errorf("failed to create new contact for user %d: %w", uid, err))
	}

	servicecontact := &Contact{}
	servicecontact.fromStorage(*cont)

	servicephones := make([]Phone, len(cont.Phones))
	for i := range cont.Phones {
		servicephones[i].fromStorage(cont.Phones[i])
	}

	return servicecontact, servicephones, nil
}

func (s *Service) GetContact(ctx context.Context, id ContactID, preload *PhonesPreload, notes bool) (*Contact, []Phone, error) {
	var pr storage.ContactPhonesPreload
	if preload != nil {
		pr = storage.ContactPhonesPreload{
			Enabled:     true,
			PrimaryOnly: preload.PrimaryOnly,
		}
	}
	cont, err := s.opts.ContactStorage.GetContact(ctx, storage.GetContactData{
		ID:       id.ID,
		UserID:   id.UserID,
		Preload:  pr,
		WithNote: notes,
	})
	if err != nil {
		err = service.TranslateStorageError(err)
		return nil, nil, s.wrapErr(fmt.Errorf("failed to get contact %d (user %d): %w", id.ID, id.UserID, err))
	}

	servicecontact := &Contact{}
	servicecontact.fromStorage(*cont)

	servicephones := make([]Phone, len(cont.Phones))
	for i := range cont.Phones {
		servicephones[i].fromStorage(cont.Phones[i])
	}

	return servicecontact, servicephones, nil
}

func (s *Service) GetContacts(ctx context.Context, uid uint, selector Selector) ([]ContactWithPhones, error) {
	if err := selector.validate(); err != nil {
		return nil, s.wrapErr(fmt.Errorf("unable to validate selector: %w", err))
	}
	conts, err := s.opts.ContactStorage.GetContacts(ctx, storage.GetContactsData{
		Selector: storage.Selector{
			Offset: selector.Offset,
			Limit:  selector.Limit,
		},
		Data: storage.GetContactData{
			UserID: uid,
			Preload: storage.ContactPhonesPreload{
				Enabled:     true,
				PrimaryOnly: true,
			},
			WithNote: false,
		},
	})
	if err != nil {
		err = service.TranslateStorageError(err)
		return nil, s.wrapErr(fmt.Errorf("failed to get contacts for user %d: %w", uid, err))
	}

	servicecontacts := make([]ContactWithPhones, len(conts))
	for i := range servicecontacts {
		servicecontacts[i].Contact.fromStorage(conts[i])
		servicecontacts[i].Phones = make([]Phone, len(conts[i].Phones))
		for j := range servicecontacts[i].Phones {
			servicecontacts[i].Phones[j].fromStorage(conts[i].Phones[j])
		}
	}

	return servicecontacts, nil
}

func (s *Service) DeleteContact(ctx context.Context, contact ContactID) error {
	err := s.opts.ContactStorage.DeleteContact(ctx, storage.DeleteContactData{
		ID:     contact.ID,
		UserID: contact.UserID,
	})
	if err != nil {
		err = service.TranslateStorageError(err)
		return s.wrapErr(fmt.Errorf("failed to delete contact %d: %w", contact.ID, err))
	}

	return nil
}
