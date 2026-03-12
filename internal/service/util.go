package service

import (
	"errors"

	"github.com/gmlazutin/comparch-lab-3mod-3/internal/storage"
)

func TranslateStorageError(err error) error {
	{
		cerr := storage.NotFoundError{}
		if errors.As(err, &cerr) {
			switch cerr.Field {
			case storage.ContactField:
				return ErrContactNotFound
			case storage.UserField:
				return ErrUserNotFound
			}
		}
	}
	{
		cerr := storage.AlreadyExistsError{}
		if errors.As(err, &cerr) {
			switch cerr.Field {
			case storage.UserField:
				return ErrUserAlreadyExists
			}
		}
	}
	{
		cerr := storage.MaxCountError{}
		if errors.As(err, &cerr) {
			switch cerr.Field {
			case storage.PhoneConstraintAllField:
				return ErrMaxPhonesCountExceeded
			case storage.PhoneConstraintPrimaryField:
				return ErrMoreThanOnePrimaryPhone
			}
		}
	}
	{
		cerr := storage.MinCountError{}
		if errors.As(err, &cerr) {
			switch cerr.Field {
			case storage.PhoneConstraintPrimaryField:
				return ErrMinimumOnePrimaryRequired
			}
		}
	}

	return err
}
