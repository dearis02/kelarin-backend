package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/volatiletech/null/v9"
)

// region repo types

type UserRole int16

const (
	UserRoleAdmin UserRole = iota + 1
	UserRoleCustomer
	UserRoleServiceProvider
)

type User struct {
	ID             uuid.UUID    `db:"id"`
	AuthProvider   AuthProvider `db:"auth_provider"`
	Role           UserRole     `db:"role"`
	Name           string       `db:"name"`
	Email          string       `db:"email"`
	Password       null.String  `db:"password"`
	IsSuspended    bool         `db:"is_suspended"`
	SuspendedCount int16        `db:"suspended_count"`
	SuspendedFrom  null.Time    `db:"suspended_from"`
	SuspendedTo    null.Time    `db:"suspended_to"`
	IsBanned       bool         `db:"is_banned"`
	BannedAt       null.Time    `db:"banned_at"`
	CreatedAt      time.Time    `db:"created_at"`
}

// end of region repo types
