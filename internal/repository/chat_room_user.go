package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ChatRoomUser interface {
	GetRoomIDAndIsExistsByUserIDs(ctx context.Context, userIDs []uuid.UUID) (uuid.UUID, bool, error)
	CreateTx(ctx context.Context, tx *sqlx.Tx, req []types.ChatRoomUser) error
}

type chatRoomUserImpl struct {
	db *sqlx.DB
}

func NewChatRoomUser(db *sqlx.DB) ChatRoomUser {
	return &chatRoomUserImpl{db: db}
}

func (r *chatRoomUserImpl) CreateTx(ctx context.Context, tx *sqlx.Tx, req []types.ChatRoomUser) error {
	query := `
		INSERT INTO chat_room_users (
			chat_room_id,
			user_id,
			created_at
		)
		VALUES (
			:chat_room_id, 
			:user_id, 
			:created_at
		)
	`

	if _, err := tx.NamedExecContext(ctx, query, req); err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *chatRoomUserImpl) GetRoomIDAndIsExistsByUserIDs(ctx context.Context, userIDs []uuid.UUID) (uuid.UUID, bool, error) {
	query := `
		SELECT
			chat_room_id
		FROM
			chat_room_users
		WHERE
			user_id IN(?)
		GROUP BY
			chat_room_id
		HAVING
			COUNT(DISTINCT user_id) = ?
	`

	var roomID uuid.UUID
	query, args, err := sqlx.In(query, userIDs, len(userIDs))
	if err != nil {
		return roomID, false, errors.New(err)
	}

	query = r.db.Rebind(query)
	err = r.db.GetContext(ctx, &roomID, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return roomID, false, nil
	} else if err != nil {
		return roomID, false, errors.New(err)
	}

	return roomID, true, nil
}
