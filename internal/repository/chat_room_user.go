package repository

import (
	"context"
	"database/sql"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ChatRoomUser interface {
	GetRoomIDAndIsExistsByUserIDs(ctx context.Context, userIDs []uuid.UUID) (uuid.UUID, bool, error)
	CreateTx(ctx context.Context, tx dbUtil.Tx, req []types.ChatRoomUser) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]types.ChatRoomUserWithServiceIDAndOfferID, error)
	FindRecipientByChatRoomIDs(ctx context.Context, userID uuid.UUID, roomIDs uuid.UUIDs) ([]types.ChatRoomUser, error)
	FindRecipientByChatRoomID(ctx context.Context, userID uuid.UUID, roomID uuid.UUID) (types.ChatRoomUser, error)
	FindByChatRoomIDAndUserID(ctx context.Context, roomID, userID uuid.UUID) (types.ChatRoomUser, error)
}

type chatRoomUserImpl struct {
	db *sqlx.DB
}

func NewChatRoomUser(db *sqlx.DB) ChatRoomUser {
	return &chatRoomUserImpl{db: db}
}

func (r *chatRoomUserImpl) CreateTx(ctx context.Context, _tx dbUtil.Tx, req []types.ChatRoomUser) error {
	tx, err := dbUtil.CastSqlxTx(_tx)
	if err != nil {
		return err
	}

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

func (r *chatRoomUserImpl) FindByUserID(ctx context.Context, userID uuid.UUID) ([]types.ChatRoomUserWithServiceIDAndOfferID, error) {
	res := []types.ChatRoomUserWithServiceIDAndOfferID{}

	query := `
		SELECT
			chat_room_users.chat_room_id,
			chat_room_users.user_id,
			chat_room_users.created_at,
			chat_rooms.service_id,
			chat_rooms.offer_id
		FROM chat_room_users
		INNER JOIN chat_rooms
			ON chat_rooms.id = chat_room_users.chat_room_id
		WHERE chat_room_users.user_id = $1
		ORDER BY id DESC
	`

	if err := r.db.SelectContext(ctx, &res, query, userID); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *chatRoomUserImpl) FindRecipientByChatRoomIDs(ctx context.Context, userID uuid.UUID, roomIDs uuid.UUIDs) ([]types.ChatRoomUser, error) {
	res := []types.ChatRoomUser{}

	query := `
		SELECT
			chat_room_id,
			user_id,
			created_at
		FROM chat_room_users
		WHERE user_id != $1
			AND chat_room_id = ANY($2)
	`

	if err := r.db.SelectContext(ctx, &res, query, userID, pq.Array(roomIDs)); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *chatRoomUserImpl) FindByChatRoomIDAndUserID(ctx context.Context, roomID, userID uuid.UUID) (types.ChatRoomUser, error) {
	res := types.ChatRoomUser{}

	query := `
		SELECT
			chat_room_id,
			user_id,
			created_at
		FROM chat_room_users
		WHERE chat_room_id = $1
			AND user_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, roomID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *chatRoomUserImpl) FindRecipientByChatRoomID(ctx context.Context, userID uuid.UUID, roomID uuid.UUID) (types.ChatRoomUser, error) {
	res := types.ChatRoomUser{}

	query := `
		SELECT
			chat_room_id,
			user_id,
			created_at
		FROM chat_room_users
		WHERE user_id != $1
			AND chat_room_id = $2
	`

	err := r.db.GetContext(ctx, &res, query, userID, roomID)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errors.New(types.ErrNoData)
	} else if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
