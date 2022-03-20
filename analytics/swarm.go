package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"teknologi-umum-bot/shared"
	"teknologi-umum-bot/utils"
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

func (d *Dependency) PurgeBots(m *tb.Message) {
	admins, err := d.Bot.AdminsOf(m.Chat)
	if err != nil {
		shared.HandleError(fmt.Errorf("get admins: %w", err), d.Logger)
		return
	}

	if !utils.IsAdmin(admins, m.Sender) {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := d.DB.Connx(ctx)
	if err != nil {
		shared.HandleError(fmt.Errorf("connection pool: %w", err), d.Logger)
		return
	}
	defer c.Close()

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: false})
	if err != nil {
		shared.HandleError(fmt.Errorf("begin transaction: %w", err), d.Logger)
		return
	}

	rows, err := tx.QueryContext(
		ctx,
		`SELECT
			user_id
		FROM
			captcha_swarm
		WHERE
			finished_captcha = false
		AND
			joined_at > NOW() - INTERVAL '1 day'`,
	)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			shared.HandleError(fmt.Errorf("rollback: %w", e), d.Logger)
			return
		}

		shared.HandleError(fmt.Errorf("query: %w", err), d.Logger)
		return
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			if e := tx.Rollback(); e != nil {
				shared.HandleError(fmt.Errorf("rollback: %w", e), d.Logger)
				return
			}

			shared.HandleError(fmt.Errorf("scan: %w", err), d.Logger)
			return
		}

		err = d.Bot.Ban(m.Chat, &tb.ChatMember{
			RestrictedUntil: tb.Forever(),
			User: &tb.User{
				ID: int(userID),
			},
		}, true)
		if err != nil {
			// TODO: do a continue loop if user was already banned
			if e := tx.Rollback(); e != nil {
				shared.HandleError(fmt.Errorf("rollback: %w", e), d.Logger)
				return
			}

			shared.HandleError(fmt.Errorf("ban: %w", err), d.Logger)
			return
		}

		userIDs = append(userIDs, userID)
		time.Sleep(time.Second * 2)
	}

	for _, userID := range userIDs {
		_, err = tx.ExecContext(
			ctx,
			`DELETE FROM
				captcha_swarm
			WHERE
				user_id = $1`,
			userID,
		)
		if err != nil {
			if e := tx.Rollback(); e != nil {
				shared.HandleError(fmt.Errorf("rollback: %w", e), d.Logger)
				return
			}

			shared.HandleError(fmt.Errorf("delete: %w", err), d.Logger)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		if e := tx.Rollback(); e != nil {
			shared.HandleError(fmt.Errorf("rollback: %w", e), d.Logger)
			return
		}

		shared.HandleError(fmt.Errorf("commit: %w", err), d.Logger)
		return
	}

	_, err = d.Bot.Send(m.Chat, fmt.Sprintf("%d bots have been banned", len(userIDs)))
	if err != nil {
		shared.HandleError(fmt.Errorf("send: %w", err), d.Logger)
		return
	}
}
