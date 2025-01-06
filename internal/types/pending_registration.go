package types

import (
	"fmt"
)

func GetPendingRegistrationKey(id string) string {
	return fmt.Sprintf("%s:%s", "pending_registration", id)
}
