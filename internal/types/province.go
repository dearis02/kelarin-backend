package types

// region repo types

type Province struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// end of region repo types

// region service types

type ProvinceGetAllRes struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// end of region service types
