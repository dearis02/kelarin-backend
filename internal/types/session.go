package types

import "fmt"

const (
	SessionKey = "session"
)

func GetSessionKey(id string) string {
	return fmt.Sprintf("%s:%s", SessionKey, id)
}
