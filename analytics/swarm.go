package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"teknologi-umum-bot/shared"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)
func (d *Dependency) SwarmLog(user *tb.User, groupID int64, finishedCaptcha bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	c, err := d.DB.Connx(ctx)
	if err != nil {
		shared.HandleError(fmt.Errorf("connection pool: %w", err), d.Logger)
		return
	}
	defer c.Close()

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted, ReadOnly: false})
	if err != nil {
		shared.HandleError(fmt.Errorf("begin transaction: %w", err), d.Logger)
		return
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO
			captcha_swarm
			(user_id, group_id, username, display_name, finished_captcha, joined_at)
		VALUES
			($1, $2, $3, $4, $5, $6)`,
		user.ID,
		groupID,
		user.Username,
		user.FirstName,
		finishedCaptcha,
		time.Now(),
	)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			shared.HandleError(fmt.Errorf("rollback: %w", e), d.Logger)
			return
		}

		shared.HandleError(fmt.Errorf("insert: %w", err), d.Logger)
		return
	}

	if err := tx.Commit(); err != nil {
		if e := tx.Rollback(); e != nil {
			shared.HandleError(fmt.Errorf("rollback: %w", e), d.Logger)
			return
		}

		shared.HandleError(fmt.Errorf("commit: %w", err), d.Logger)
		return
	}
}

func (d *Dependency) UpdateSwarm(user *tb.User, groupID int64, finishedCaptcha bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	c, err := d.DB.Connx(ctx)
	if err != nil {
		shared.HandleError(fmt.Errorf("connection pool: %w", err), d.Logger)
		return
	}
	defer c.Close()

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted, ReadOnly: false})
	if err != nil {
		shared.HandleError(fmt.Errorf("begin transaction: %w", err), d.Logger)
		return
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE
			captcha_swarm
		SET
			finished_captcha = $1
		WHERE
			user_id = $2
		AND
			group_id = $3`,
		finishedCaptcha,
		user.ID,
		groupID,
	)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			shared.HandleError(fmt.Errorf("rollback: %w", e), d.Logger)
			return
		}

		shared.HandleError(fmt.Errorf("insert: %w", err), d.Logger)
		return
	}

	if err := tx.Commit(); err != nil {
		if e := tx.Rollback(); e != nil {
			shared.HandleError(fmt.Errorf("rollback: %w", e), d.Logger)
			return
		}

		shared.HandleError(fmt.Errorf("commit: %w", err), d.Logger)
		return
	}
}
