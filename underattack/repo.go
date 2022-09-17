package underattack

import (
	"context"
	"database/sql"
	"errors"
	"teknologi-umum-bot/shared"
	"time"
)

// GetUnderAttackEntry will acquire under attack entry for specified groupID.
func (d *Dependency) GetUnderAttackEntry(ctx context.Context, groupID int64) (underattack, error) {
	c, err := d.DB.Connx(ctx)
	if err != nil {
		return underattack{}, err
	}
	defer func() {
		err := c.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			shared.HandleError(err, d.Logger)
		}
	}()

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true})
	if err != nil {
		return underattack{}, err
	}

	var entry underattack

	err = tx.QueryRowxContext(
		ctx,
		"SELECT * FROM under_attack WHERE group_id = $1 ORDER BY updated_at DESC",
		groupID,
	).StructScan(&entry)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return underattack{}, e
		}

		if errors.Is(err, sql.ErrNoRows) {
			go func(groupID int64) {
				time.Sleep(time.Second * 5)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
				defer cancel()

				err := d.CreateNewEntry(ctx, groupID)
				if err != nil {
					shared.HandleError(err, d.Logger)
				}
			}(groupID)

			return underattack{}, nil
		}

		return underattack{}, err
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return underattack{}, e
		}

		return underattack{}, err
	}

	return entry, nil
}

// CreateNewEntry will create a new entry for given groupID.
// This should only be executed if the group entry does not exists on the database.
// If it already exists, it will do nothing.
func (d *Dependency) CreateNewEntry(ctx context.Context, groupID int64) error {
	c, err := d.DB.Connx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err := c.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			shared.HandleError(err, d.Logger)
		}
	}()

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO
			under_attack
			(group_id, is_under_attack, expires_at, notification_message_id, updated_at)
		VALUES
			($1, $2, $3, $4, $5)
		ON CONFLICT (group_id)
		DO NOTHING`,
		groupID,
		false,
		time.Time{},
		0,
		time.Now(),
	)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	return nil
}

// SetUnderAttackStatus will update the given groupID entry to the given parameters.
// If the groupID entry does not exists, it will create a new one.
func (d *Dependency) SetUnderAttackStatus(ctx context.Context, groupID int64, underAttack bool, expiresAt time.Time, notificationMessageID int64) error {
	c, err := d.DB.Connx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err := c.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			shared.HandleError(err, d.Logger)
		}
	}()

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO
			under_attack
			(group_id, is_under_attack, expires_at, notification_message_id, updated_at)
		VALUES
			($1, $2, $3, $4, $5)
		ON CONFLICT (group_id)
		DO UPDATE
		SET
			is_under_attack = $2,
			expires_at = $3,
			notification_message_id = $4,
			updated_at = $5`,
		groupID,
		underAttack,
		expiresAt,
		notificationMessageID,
		time.Now(),
	)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	return nil
}
