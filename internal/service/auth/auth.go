package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gmlazutin/comparch-lab-3mod-3/internal/logging"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/auth/session"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Login       string
	SessionData *session.Session
}

func (u *User) fromStorage(user storage.User) {
	u.Login = user.Login
}

type Options struct {
	Storage                   storage.UserStorage
	ServiceOpts               service.ServiceOptions
	SessionValidatorGenerator session.AuthTokenValidatorGenerator
	SessionExpireTimeout      time.Duration
}

type Service struct {
	opts Options
}

func New(options Options) *Service {
	if options.SessionExpireTimeout == 0 {
		options.SessionExpireTimeout = time.Hour
	}
	if options.ServiceOpts.Logger == nil {
		options.ServiceOpts.Logger = logging.EmptyLogger()
	}

	return &Service{
		opts: options,
	}
}

func (s *Service) wrapErr(err error) error {
	return fmt.Errorf("authService: %w", err)
}

//todo: take it into separate PasswordHashProvider interface

func (s *Service) genPasswordHashBcrypt(password string) (storage.UserPassword, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return storage.UserPassword{
		Hash: string(bytes),
		Algo: "bcrypt",
	}, err
}

func (s *Service) checkPassword(usrpass storage.UserPassword, password string) error {
	switch usrpass.Algo {
	case "bcrypt":
		err := bcrypt.CompareHashAndPassword([]byte(usrpass.Hash), []byte(password))
		return err
	default:
		return service.ErrIncorrectPassword
	}
}

func (s *Service) CreateUserSimple(ctx context.Context, login, password string, ts time.Time) (*session.Session, string, error) {
	usrpass, err := s.genPasswordHashBcrypt(password)
	if err != nil {
		return nil, "", s.wrapErr(fmt.Errorf("failed to hash password for new user %q: %w", login, err))
	}
	usr, err := s.opts.Storage.AddUser(ctx, storage.AddUserData{
		User: storage.User{
			Login:    login,
			Password: usrpass,
		},
	})
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			err = service.ErrUserAlreadyExists
		}

		return nil, "", s.wrapErr(fmt.Errorf("failed to create user with login %q: %w", login, err))
	}
	sess := &session.Session{
		UserID:  usr.ID,
		Expires: ts.Add(s.opts.SessionExpireTimeout),
	}
	tkn, err := s.opts.SessionValidatorGenerator.Generate(ctx, *sess)
	if err != nil {
		return nil, "", s.wrapErr(fmt.Errorf("failed to create new session for created user %d: %w", usr.ID, err))
	}

	return sess, tkn, nil
}

func (s *Service) AuthUserByPassword(ctx context.Context, login, password string, ts time.Time) (*session.Session, string, error) {
	user, err := s.opts.Storage.GetUser(ctx, storage.GetUserData{
		Login:           login,
		WithCredentials: true,
	})
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			err = service.ErrUserNotFound
		}

		return nil, "", s.wrapErr(fmt.Errorf("failed to authenticate user with login %q: %w", login, err))
	}

	if s.checkPassword(user.Password, password) != nil {
		return nil, "", s.wrapErr(fmt.Errorf("failed to authenticate user %d: %w", user.ID, service.ErrIncorrectPassword))
	}
	sess := &session.Session{
		UserID:  user.ID,
		Expires: ts.Add(s.opts.SessionExpireTimeout),
	}
	tkn, err := s.opts.SessionValidatorGenerator.Generate(ctx, *sess)
	if err != nil {
		return nil, "", s.wrapErr(fmt.Errorf("failed to create new session for authenticated user %d: %w", user.ID, err))
	}

	return sess, tkn, nil
}

func (s *Service) CheckUserSession(ctx context.Context, session string, ts time.Time) (*session.Session, error) {
	sess, err := s.opts.SessionValidatorGenerator.Validate(ctx, session, ts)
	if err != nil {
		return nil, s.wrapErr(fmt.Errorf("failed to validate session: %w", err))
	}

	return sess, nil
}

func (s *Service) GetUserBySession(ctx context.Context, session *session.Session) (*User, error) {
	user, err := s.opts.Storage.GetUser(ctx, storage.GetUserData{
		ID:              session.UserID,
		WithCredentials: false,
	})
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			err = service.ErrUserNotFound
		}

		return nil, s.wrapErr(fmt.Errorf("failed to get user %d: %w", session.UserID, err))
	}

	usr := &User{}
	usr.fromStorage(*user)
	usr.SessionData = session

	return usr, nil
}
