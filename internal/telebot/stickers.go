package telebot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type (
	StickerSetType   = string
	StickerSetFormat = string
	MaskFeature      = string
)

const (
	StickerRegular     StickerSetType = "regular"
	StickerMask        StickerSetType = "mask"
	StickerCustomEmoji StickerSetType = "custom_emoji"
)

const (
	StickerStatic   StickerSetFormat = "static"
	StickerAnimated StickerSetFormat = "animated"
	StickerVideo    StickerSetFormat = "video"
)

const (
	MaskForehead MaskFeature = "forehead"
	MaskEyes     MaskFeature = "eyes"
	MaskMouth    MaskFeature = "mouth"
	MaskChin     MaskFeature = "chin"
)

// StickerSet represents a sticker set.
type StickerSet struct {
	Type          StickerSetType   `json:"sticker_type"`
	Format        StickerSetFormat `json:"sticker_format"`
	Name          string           `json:"name"`
	Title         string           `json:"title"`
	Stickers      []Sticker        `json:"stickers"`
	Thumbnail     *Photo           `json:"thumbnail"`
	Emojis        string           `json:"emojis"`
	ContainsMasks bool             `json:"contains_masks"` // FIXME: can be removed
	MaskPosition  *MaskPosition    `json:"mask_position"`
	Repaint       bool             `json:"needs_repainting"`

	// Input is a field used in createNewStickerSet method to specify a list
	// of pre-defined stickers of type InputSticker to add to the set.
	Input []InputSticker
}

type InputSticker struct {
	File
	Sticker      string        `json:"sticker"`
	Format       string        `json:"format"`
	MaskPosition *MaskPosition `json:"mask_position"`
	Emojis       []string      `json:"emoji_list"`
	Keywords     []string      `json:"keywords"`
}

// MaskPosition describes the position on faces where
// a mask should be placed by default.
type MaskPosition struct {
	Feature MaskFeature `json:"point"`
	XShift  float32     `json:"x_shift"`
	YShift  float32     `json:"y_shift"`
	Scale   float32     `json:"scale"`
}

// UploadSticker uploads a sticker file for later use.
func (b *Bot) UploadSticker(ctx context.Context, to Recipient, format StickerSetFormat, f File) (*File, error) {
	params := map[string]string{
		"user_id":        to.Recipient(),
		"sticker_format": format,
	}

	data, err := b.sendFiles(ctx, "uploadStickerFile", map[string]File{"0": f}, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result File
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, wrapError(err)
	}
	return &resp.Result, nil
}

// StickerSet returns a sticker set on success.
func (b *Bot) StickerSet(ctx context.Context, name string) (*StickerSet, error) {
	data, err := b.Raw(ctx, "getStickerSet", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result *StickerSet
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, wrapError(err)
	}
	return resp.Result, nil
}

// CreateStickerSet creates a new sticker set.
func (b *Bot) CreateStickerSet(ctx context.Context, of Recipient, set *StickerSet) error {
	files := make(map[string]File)
	for i, s := range set.Input {
		repr := s.File.process(strconv.Itoa(i), files)
		if repr == "" {
			return fmt.Errorf("telebot: sticker #%d does not exist", i+1)
		}
		set.Input[i].Sticker = repr
	}

	data, _ := json.Marshal(set.Input)

	params := map[string]string{
		"user_id":  of.Recipient(),
		"name":     set.Name,
		"title":    set.Title,
		"stickers": string(data),
	}
	if set.Type != "" {
		params["sticker_type"] = set.Type
	}
	if set.Repaint {
		params["needs_repainting"] = "true"
	}

	_, err := b.sendFiles(ctx, "createNewStickerSet", files, params)
	return err
}

// AddStickerToSet adds a new sticker to the existing sticker set.
func (b *Bot) AddStickerToSet(ctx context.Context, of Recipient, name string, sticker InputSticker) error {
	files := make(map[string]File)
	repr := sticker.File.process("0", files)
	if repr == "" {
		return errors.New("telebot: sticker does not exist")
	}

	sticker.Sticker = repr
	data, _ := json.Marshal(sticker)

	params := map[string]string{
		"user_id": of.Recipient(),
		"name":    name,
		"sticker": string(data),
	}

	_, err := b.sendFiles(ctx, "addStickerToSet", files, params)
	return err
}

// SetStickerPosition moves a sticker in set to a specific position.
func (b *Bot) SetStickerPosition(ctx context.Context, sticker string, position int) error {
	params := map[string]string{
		"sticker":  sticker,
		"position": strconv.Itoa(position),
	}

	_, err := b.Raw(ctx, "setStickerPositionInSet", params)
	return err
}

// DeleteSticker deletes a sticker from a set created by the bot.
func (b *Bot) DeleteSticker(ctx context.Context, sticker string) error {
	_, err := b.Raw(ctx, "deleteStickerFromSet", map[string]string{"sticker": sticker})
	return err

}

// SetStickerSetThumb sets a thumbnail of the sticker set.
// Animated thumbnails can be set for animated sticker sets only.
//
// Thumbnail must be a PNG image, up to 128 kilobytes in size
// and have width and height exactly 100px, or a TGS animation
// up to 32 kilobytes in size.
//
// Animated sticker set thumbnail can't be uploaded via HTTP URL.
func (b *Bot) SetStickerSetThumb(ctx context.Context, of Recipient, set *StickerSet) error {
	if set.Thumbnail == nil {
		return errors.New("telebot: thumbnail is required")
	}

	files := make(map[string]File)
	repr := set.Thumbnail.File.process("thumb", files)
	if repr == "" {
		return errors.New("telebot: thumbnail does not exist")
	}

	params := map[string]string{
		"user_id":   of.Recipient(),
		"name":      set.Name,
		"format":    set.Format,
		"thumbnail": repr,
	}

	_, err := b.sendFiles(ctx, "setStickerSetThumbnail", files, params)
	return err
}

// SetStickerSetTitle sets the title of a created sticker set.
func (b *Bot) SetStickerSetTitle(ctx context.Context, s StickerSet) error {
	params := map[string]string{
		"name":  s.Name,
		"title": s.Title,
	}

	_, err := b.Raw(ctx, "setStickerSetTitle", params)
	return err
}

// DeleteStickerSet deletes a sticker set that was created by the bot.
func (b *Bot) DeleteStickerSet(ctx context.Context, name string) error {
	params := map[string]string{"name": name}

	_, err := b.Raw(ctx, "deleteStickerSet", params)
	return err
}

// SetStickerEmojis changes the list of emoji assigned to a regular or custom emoji sticker.
func (b *Bot) SetStickerEmojis(ctx context.Context, sticker string, emojis []string) error {
	data, err := json.Marshal(emojis)
	if err != nil {
		return err
	}

	params := map[string]string{
		"sticker":    sticker,
		"emoji_list": string(data),
	}

	_, err = b.Raw(ctx, "setStickerEmojiList", params)
	return err
}

// SetStickerKeywords changes search keywords assigned to a regular or custom emoji sticker.
func (b *Bot) SetStickerKeywords(ctx context.Context, sticker string, keywords []string) error {
	mk, err := json.Marshal(keywords)
	if err != nil {
		return err
	}

	params := map[string]string{
		"sticker":  sticker,
		"keywords": string(mk),
	}

	_, err = b.Raw(ctx, "setStickerKeywords", params)
	return err
}

// SetStickerMaskPosition changes the mask position of a mask sticker.
func (b *Bot) SetStickerMaskPosition(ctx context.Context, sticker string, mask MaskPosition) error {
	data, err := json.Marshal(mask)
	if err != nil {
		return err
	}

	params := map[string]string{
		"sticker":       sticker,
		"mask_position": string(data),
	}

	_, err = b.Raw(ctx, "setStickerMaskPosition", params)
	return err
}

// CustomEmojiStickers returns the information about custom emoji stickers by their ids.
func (b *Bot) CustomEmojiStickers(ctx context.Context, ids []string) ([]Sticker, error) {
	data, _ := json.Marshal(ids)

	params := map[string]string{
		"custom_emoji_ids": string(data),
	}

	data, err := b.Raw(ctx, "getCustomEmojiStickers", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result []Sticker
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, wrapError(err)
	}
	return resp.Result, nil
}

// SetCustomEmojiStickerSetThumb sets the thumbnail of a custom emoji sticker set.
func (b *Bot) SetCustomEmojiStickerSetThumb(ctx context.Context, name, id string) error {
	params := map[string]string{
		"name":            name,
		"custom_emoji_id": id,
	}

	_, err := b.Raw(ctx, "setCustomEmojiStickerSetThumbnail", params)
	return err
}

// ReplaceStickerInSet returns True on success, if existing sticker was replaced with a new one.
func (b *Bot) ReplaceStickerInSet(ctx context.Context, of Recipient, stickerSet, oldSticker string, sticker InputSticker) (bool, error) {
	files := make(map[string]File)

	repr := sticker.File.process("0", files)
	if repr == "" {
		return false, errors.New("telebot: sticker does not exist")
	}
	sticker.Sticker = repr

	data, err := json.Marshal(sticker)
	if err != nil {
		return false, err
	}

	params := map[string]string{
		"user_id":     of.Recipient(),
		"name":        stickerSet,
		"old_sticker": oldSticker,
		"sticker":     string(data),
	}

	_, err = b.sendFiles(ctx, "replaceStickerInSet", files, params)
	return true, err
}
