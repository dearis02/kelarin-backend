package types

// region repo types

type City struct {
	ID         int64  `db:"id"`
	ProvinceID int64  `db:"province_id"`
	Name       string `db:"name"`
}

// end of region repo types
