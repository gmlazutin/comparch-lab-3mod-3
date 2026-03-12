package gorm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gmlazutin/comparch-lab-3mod-3/internal/logging"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/storage"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/storage/gorm/model"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

//todo: rewrite to pure pgsql driver for better perfomance
//and to avoid dirty GORM sql hacks.

// implements ContactStorage, NumberStorage, UserStorage
type DB struct {
	db   *gorm.DB
	opts Options
}

type Options struct {
	Driver string
	Dsn    string
	Opts   storage.Options
}

func New(opts Options) (*DB, error) {
	if opts.Opts.Logger == nil {
		opts.Opts.Logger = logging.EmptyLogger()
	}

	var dialector gorm.Dialector
	switch opts.Driver {
	case "pgsql":
		dialector = postgres.Open(opts.Dsn)
	case "sqlite":
		dialector = sqlite.Open(opts.Dsn)
	default:
		return nil, fmt.Errorf("gormStorageProvider: unknown db driver: %s", opts.Driver)
	}

	opts.Opts.Logger = opts.Opts.Logger.With(logging.Service("gormStorageProvider"))

	//todo: write custom slog logger for GORM. This is workaround
	//due to default behavior of GORM slogLogger
	var log logger.Interface
	if opts.Opts.Logger.Enabled(nil, slog.LevelDebug) {
		log = logger.NewSlogLogger(opts.Opts.Logger, logger.Config{
			LogLevel: logger.Info,
		})
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		TranslateError: true,
		Logger:         log,
	})
	if err != nil {
		return nil, fmt.Errorf("gormStorageProvider: failed to establish db conn: %w", err)
	}

	return &DB{
		db:   db,
		opts: opts,
	}, nil
}

func (db *DB) PerfomMigrations(ctx context.Context) error {
	tx := db.db.WithContext(ctx)
	if err := tx.AutoMigrate(&model.Phone{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&model.Contact{}); err != nil {
		return err
	}
	if err := tx.AutoMigrate(&model.User{}); err != nil {
		return err
	}

	return nil
}

func (db *DB) Stop() error {
	sqldb, err := db.db.DB()
	if err != nil {
		return fmt.Errorf("gormStorageProvider: failed to get sqldb instance: %w", err)
	}
	err = sqldb.Close()
	if err != nil {
		return fmt.Errorf("gormStorageProvider: failed to close sqldb: %w", err)
	}

	return nil
}

func (db *DB) translateError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return storage.ErrNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return storage.ErrAlreadyExists
	}
	return err
}

//UserStorage

func (db *DB) AddUser(ctx context.Context, data storage.AddUserData) (*storage.User, error) {
	user := &model.User{}
	user.FromUser(data.User)
	if tx := db.db.WithContext(ctx).Create(user); tx.Error != nil {
		return nil, db.translateError(tx.Error)
	}
	return user.ToUser(), nil
}

func (db *DB) GetUser(ctx context.Context, data storage.GetUserData) (*storage.User, error) {
	var user model.User
	selected := []string{"id", "login"}
	if data.WithCredentials {
		selected = append(selected, "password_hash", "password_algo")
	}
	var qargs []any
	if data.ID > 0 {
		qargs = []any{data.ID}
	} else if len(data.Login) > 0 {
		qargs = []any{"login = ?", data.Login}
	} else {
		return nil, storage.ErrNotFound
	}
	if tx := db.db.WithContext(ctx).Select(selected).First(&user, qargs...); tx.Error != nil {
		return nil, db.translateError(tx.Error)
	}
	return user.ToUser(), nil
}

//ContactStorage

func (db *DB) AddContact(ctx context.Context, data storage.AddContactData) (*storage.Contact, error) {
	contact := &model.Contact{}
	contact.FromContact(data.Contact)
	var phones []storage.Phone
	err := db.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if tx := tx.Create(contact); tx.Error != nil {
			return tx.Error
		}

		dbphones := make([]model.Phone, len(data.InitialPhones))
		for i := range dbphones {
			dbphones[i].FromPhone(data.InitialPhones[i])
		}
		err := db.addPhonesBatch(tx, contact.UserID, contact.ID, false, dbphones)
		if err != nil {
			return err
		}
		phones = make([]storage.Phone, len(dbphones))
		for i := range dbphones {
			phones[i] = *dbphones[i].ToPhone()
		}

		return nil
	})
	if err != nil {
		return nil, db.translateError(err)
	}
	card := contact.ToContact()
	card.Phones = phones
	return card, nil
}

// Warning:
// This can break some statements due to special processing of zero-valued offset/limit!
// Check Selector Limit/Offset values first!
func (db *DB) applySelectorRules(tx *gorm.DB, selector storage.Selector) *gorm.DB {
	if selector.Limit > 0 {
		tx = tx.Limit(int(selector.Limit))
	}
	if selector.Offset > 0 {
		tx = tx.Offset(int(selector.Offset))
	}
	return tx
}

func (db *DB) getContacts(ctx context.Context, data storage.GetContactsData) ([]storage.Contact, error) {
	var cards []model.Contact
	tx := db.db.WithContext(ctx)
	if data.Data.Preload.Enabled {
		var prargs []any
		//todo: unoptimized query, consider JOIN instead of Preload
		if data.Data.Preload.PrimaryOnly {
			prargs = []any{"`primary` = ?", true}
		}
		tx = tx.Preload("Phones", prargs...)
	}
	tx = db.applySelectorRules(tx, data.Selector)
	fields := []string{"id", "user_id", "name", "birthday"}
	if data.Data.WithNote {
		fields = append(fields, "note")
	}
	where := &model.Contact{UserID: data.Data.UserID}
	where.ID = data.Data.ID
	tx = tx.Where(where).Select(fields)
	tx = tx.Order(clause.OrderByColumn{
		Column: clause.Column{Table: clause.CurrentTable, Name: clause.PrimaryKey},
	})
	if tx = tx.Find(&cards); tx.Error != nil {
		return nil, db.translateError(tx.Error)
	}
	var stcards = make([]storage.Contact, len(cards))
	for i := 0; i < len(cards); i++ {
		stcards[i] = *cards[i].ToContact()
		stcards[i].Phones = make([]storage.Phone, len(cards[i].Phones))
		for j := 0; j < len(stcards[i].Phones); j++ {
			stcards[i].Phones[j] = *cards[i].Phones[j].ToPhone()
		}
	}
	return stcards, nil
}

func (db *DB) GetContacts(ctx context.Context, data storage.GetContactsData) ([]storage.Contact, error) {
	data.Data.ID = 0
	return db.getContacts(ctx, data)
}

func (db *DB) GetContact(ctx context.Context, data storage.GetContactData) (*storage.Contact, error) {
	contacts, err := db.getContacts(ctx, storage.GetContactsData{
		Selector: storage.Selector{
			Offset: 0,
			Limit:  1,
		},
		Data: data,
	})
	if err != nil {
		return nil, err
	}
	if len(contacts) == 0 {
		return nil, storage.ErrNotFound
	}

	return &contacts[0], nil
}

func (db *DB) DeleteContact(ctx context.Context, data storage.DeleteContactData) error {
	tx := db.db.WithContext(ctx).
		Where(&model.Contact{UserID: data.UserID}).
		Delete(&model.Contact{}, data.ID)
	if tx.Error != nil {
		return db.translateError(tx.Error)
	}
	//soft-delete case
	if tx.RowsAffected == 0 {
		return storage.ErrNotFound
	}

	return nil
}

func (db *DB) addPhonesBatch(tx *gorm.DB, uid uint, cid uint, check bool, phones []model.Phone) error {
	if len(phones) == 0 {
		return nil
	}

	if check {
		var dummy struct{}
		where := &model.Contact{UserID: uid}
		where.ID = cid
		err := tx.Model(&model.Contact{}).
			Select("1").
			Where(where).
			Take(&dummy).Error

		if err != nil {
			return db.translateError(err)
		}
	}

	for i := range phones {
		phones[i].ContactID = cid
	}

	return db.translateError(tx.Create(&phones).Error)
}
