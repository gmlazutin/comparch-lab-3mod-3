package service

type AuthError struct {
	message string
}

func (e AuthError) Error() string {
	return e.message
}

type ContactsError struct {
	message string
}

func (e ContactsError) Error() string {
	return e.message
}

type CustomValidationError struct {
	message string
}

func (e CustomValidationError) Error() string {
	return e.message
}

var (
	ErrIncorrectPassword = AuthError{"incorrect password"}
	ErrUserAlreadyExists = AuthError{"user already exists"}
	ErrUserNotFound      = AuthError{"user not found"}

	ErrInvalidToken = AuthError{"invalid token"}

	ErrContactNotFound = ContactsError{"contact not found"}

	ErrIncorrectSelectorValues = CustomValidationError{"incorrect selector"}
	ErrIncorrectPhone          = CustomValidationError{"incorrect phone number"}
	ErrIncorrectBirthday       = CustomValidationError{"incorrect birthday date"}
)
