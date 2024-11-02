package types

const TIME_FORMAT_TZ = "2006-01-02 15:04:05 -0700"

type ErrValidationRes struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type AppErr struct {
	Code    int
	Message string
	Err     error
}

func (e AppErr) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return e.Message
}
