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

// implements ContactStorage, NumberStorage, UserStorage
type DB struct {
	db   *gorm.DB
	opts Options
}

const (
	DB_FLUSH_DEFAULT_UNIT = 2000 //rows
)

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

func (db *DB) PerformFlush(ctx context.Context, batchSize int) error {
	for _, v := range []any{&model.User{}, &model.Phone{}, &model.Contact{}} {
		var rowsTotal int64
		for {
			sub := db.db.WithContext(ctx).
				Unscoped().
				Model(v).
				Select("id").
				Where("deleted_at IS NOT NULL").
				Limit(batchSize)
			res := db.db.WithContext(ctx).
				Unscoped().
				Where("id IN (?)", sub).
				Delete(v)

			if res.Error != nil {
				db.opts.Opts.Logger.Error("db flush error", logging.Error(res.Error), slog.String("model", fmt.Sprintf("%T", v)))
				return res.Error
			}

			if res.RowsAffected == 0 {
				break
			}

			rowsTotal += res.RowsAffected
		}

		db.opts.Opts.Logger.Debug("db flush perfomed", slog.Int64("rowsAffected", rowsTotal), slog.String("model", fmt.Sprintf("%T", v)))
	}
	return nil
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

func (db *DB) translateError(err error, field string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return storage.NotFoundError{
			Field: field,
		}
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return storage.AlreadyExistsError{
			Field: field,
		}
	}
	return fmt.Errorf("gormStorageProvider: unknown error for %q: %w", field, err)
}

//UserStorage

func (db *DB) AddUser(ctx context.Context, data storage.AddUserData) (*storage.User, error) {
	user := &model.User{}
	user.FromUser(data.User)
	if tx := db.db.WithContext(ctx).Create(user); tx.Error != nil {
		return nil, db.translateError(tx.Error, storage.UserField)
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
		return nil, storage.NotFoundError{
			Field: storage.UserField,
		}
	}
	if tx := db.db.WithContext(ctx).Select(selected).First(&user, qargs...); tx.Error != nil {
		return nil, db.translateError(tx.Error, storage.UserField)
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
			return db.translateError(tx.Error, storage.ContactField)
		}

		if len(data.InitialPhones) > int(data.PhoneConstraints.MaxAllowed) {
			return storage.MaxCountError{
				Field: storage.PhoneConstraintAllField,
			}
		}
		dbphones := make([]model.Phone, len(data.InitialPhones))
		var primaries uint
		for i := range dbphones {
			dbphones[i].FromPhone(data.InitialPhones[i])
			if dbphones[i].Primary {
				primaries++
				if primaries > data.PhoneConstraints.MaxPrimaries {
					return storage.MaxCountError{
						Field: storage.PhoneConstraintPrimaryField,
					}
				}
			}
		}
		if primaries < data.PhoneConstraints.MinPrimaries {
			return storage.MinCountError{
				Field: storage.PhoneConstraintPrimaryField,
			}
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
		return nil, err
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
		if data.Data.Preload.PrimaryOnly {
			prargs = []any{"\"primary\" = ?", true}
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
		return nil, db.translateError(tx.Error, storage.ContactField)
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
		return nil, storage.NotFoundError{
			Field: storage.ContactField,
		}
	}

	return &contacts[0], nil
}

func (db *DB) DeleteContact(ctx context.Context, data storage.DeleteContactData) error {
	tx := db.db.WithContext(ctx).
		Where(&model.Contact{UserID: data.UserID}).
		Delete(&model.Contact{}, data.ID)
	if tx.Error != nil {
		return db.translateError(tx.Error, storage.ContactField)
	}
	//soft-delete case
	if tx.RowsAffected == 0 {
		return storage.NotFoundError{
			Field: storage.ContactField,
		}
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
			return db.translateError(err, storage.ContactField)
		}
	}

	for i := range phones {
		phones[i].ContactID = cid
	}

	err := tx.Create(&phones).Error
	if err != nil {
		return db.translateError(err, storage.PhoneField)
	}
	return nil
}
