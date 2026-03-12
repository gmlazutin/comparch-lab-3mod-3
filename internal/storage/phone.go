package storage

const (
	PhoneField string = "phone"
)

type Phone struct {
	ID        uint
	ContactID uint
	Phone     string
	Primary   bool
}

const (
	PhoneConstraintAllField     string = "phone-all-count"
	PhoneConstraintPrimaryField string = "phone-primary-count"
)

type PhoneConstraints struct {
	MaxAllowed   uint
	MaxPrimaries uint
	MinPrimaries uint
}

/*type AddNumberData struct {
	//ID will be overwritten
	Phone  Phone
	UserID uint
}

type GetNumberData struct {
	ID     uint
	UserID uint
}

type DeleteNumberData struct {
	ID     uint
	UserID uint
}
*/

type PhoneStorage interface {
	//AddNumber(ctx context.Context, data AddNumberData) (*Phone, error)
	//GetNumbers(ctx context.Context, data GetNumberData) ([]Phone, error)
	//UpdateNumber(ctx context.Context, cid uint, data NumberUpdateData) (*PhoneCard, error)
	//DeleteNumber(ctx context.Context, data DeleteNumberData) error
}
