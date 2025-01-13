package types

// regin repo types

type District struct {
	ID     int64  `db:"id"`
	CityID int64  `db:"city_id"`
	Name   string `db:"name"`
}

// end of region repo types
