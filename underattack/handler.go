package underattack

import (
	"context"
	"strconv"
	"teknologi-umum-bot/shared"
	"teknologi-umum-bot/utils"
	"time"

	tb "gopkg.in/telebot.v3"
)

// EnableUnderAttackModeHandler provides a handler for /underattack command.
func (d *Dependency) EnableUnderAttackModeHandler(c tb.Context) error {
	if c.Message().Private() || c.Sender().IsBot {
		return nil
	}

	admins, err := c.Bot().AdminsOf(c.Chat())
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	if !utils.IsAdmin(admins, c.Sender()) {
		_, err := c.Bot().Send(
			c.Chat(),
			"Cuma admin yang boleh jalanin command ini. Ada baiknya kamu ping adminnya langsung :)",
			&tb.SendOptions{
				ReplyTo:           c.Message(),
				AllowWithoutReply: true,
			},
		)
		if err != nil {
			shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
			return nil
		}

		return nil
	}

	// Sender must be an admin here.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()

	// Check if we are on the under attack mode right now.
	underAttackModeEnabled, err := d.AreWe(ctx, c.Chat().ID)
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	if underAttackModeEnabled {
		_, err := c.Bot().Send(
			c.Chat(),
			"Mode under attack sudah menyala. Untuk mematikan, kirim /disableunderattack",
			&tb.SendOptions{
				ReplyTo:           c.Message(),
				AllowWithoutReply: true,
			},
		)
		if err != nil {
			shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
			return nil
		}

		return nil
	}

	expiresAt := time.Now().Add(time.Minute * 30)

	notificationMessage, err := c.Bot().Send(
		c.Chat(),
		"Grup ini dalam kondisi under attack sampai pukul "+
			expiresAt.In(time.FixedZone("WIB", 7*60*60)).Format("15:04 MST")+
			". Semua yang baru masuk ke grup ini akan langsung di ban selamanya."+
			"Untuk bisa bergabung, tunggu sampai under attack mode berakhir, atau hubungi admin grup.\n\n",
		"This group is in under attack mode until "+
			expiresAt.In(time.FixedZone("UTC +7", 7*60*60)).Format("15:04 MST")+
			". Everyone that is joining this group will be banned forever."+
			"To be able to join, wait until the under attack mode is over, or contact the group's administrator.\n\n",
		&tb.SendOptions{
			ParseMode: tb.ModeDefault,
		},
	)
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	err = d.SetUnderAttackStatus(ctx, c.Chat().ID, true, time.Now().Add(time.Minute*30), int64(notificationMessage.ID))
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	err = d.Memory.Delete("underattack:" + strconv.FormatInt(c.Chat().ID, 10))
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	err = c.Bot().Pin(notificationMessage)
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	return nil
}

// DisableUnderAttackModeHandler provides a handler for /disableunderattack command.
func (d *Dependency) DisableUnderAttackModeHandler(c tb.Context) error {
	if c.Message().Private() || c.Sender().IsBot {
		return nil
	}

	admins, err := c.Bot().AdminsOf(c.Chat())
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	if !utils.IsAdmin(admins, c.Sender()) {
		_, err := c.Bot().Send(
			c.Chat(),
			"Cuma admin yang boleh jalanin command ini. Ada baiknya kamu ping adminnya langsung :)",
			&tb.SendOptions{
				ReplyTo:           c.Message(),
				AllowWithoutReply: true,
			},
		)
		if err != nil {
			shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
			return nil
		}

		return nil
	}

	// Sender must be an admin here.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()

	// Check if we are on the under attack mode right now.
	underAttackModeEnabled, err := d.AreWe(ctx, c.Chat().ID)
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	if !underAttackModeEnabled {
		return nil
	}

	underAttackEntry, err := d.GetUnderAttackEntry(ctx, c.Chat().ID)
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	err = d.SetUnderAttackStatus(ctx, c.Chat().ID, false, time.Now(), 0)
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	err = d.Memory.Delete("underattack:" + strconv.FormatInt(c.Chat().ID, 10))
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	err = c.Bot().Unpin(c.Chat(), int(underAttackEntry.NotificationMessageID))
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, c.Message())
		return nil
	}

	return nil
}
