package telebot

import (
	"context"
	"io"
)

// API is the interface that wraps all basic methods for interacting
// with Telegram Bot API.
type API interface {
	Raw(ctx context.Context, method string, payload interface{}) ([]byte, error)

	Accept(ctx context.Context, query *PreCheckoutQuery, errorMessage ...string) error
	AddStickerToSet(ctx context.Context, of Recipient, name string, sticker InputSticker) error
	AdminsOf(ctx context.Context, chat *Chat) ([]ChatMember, error)
	Answer(ctx context.Context, query *Query, resp *QueryResponse) error
	AnswerWebApp(ctx context.Context, query *Query, r Result) (*WebAppMessage, error)
	ApproveJoinRequest(ctx context.Context, chat Recipient, user *User) error
	Ban(ctx context.Context, chat *Chat, member *ChatMember, revokeMessages ...bool) error
	BanSenderChat(ctx context.Context, chat *Chat, sender Recipient) error
	BusinessConnection(ctx context.Context, d string) (*BusinessConnection, error)
	ChatByID(ctx context.Context, id int64) (*Chat, error)
	ChatByUsername(ctx context.Context, name string) (*Chat, error)
	ChatMemberOf(ctx context.Context, chat, user Recipient) (*ChatMember, error)
	Close(ctx context.Context) (bool, error)
	CloseGeneralTopic(ctx context.Context, chat *Chat) error
	CloseTopic(ctx context.Context, chat *Chat, topic *Topic) error
	Commands(ctx context.Context, opts ...interface{}) ([]Command, error)
	Copy(ctx context.Context, to Recipient, msg Editable, opts ...interface{}) (*Message, error)
	CopyMany(ctx context.Context, to Recipient, msgs []Editable, opts ...*SendOptions) ([]Message, error)
	CreateInviteLink(ctx context.Context, chat Recipient, link *ChatInviteLink) (*ChatInviteLink, error)
	CreateInvoiceLink(ctx context.Context, i Invoice) (string, error)
	CreateStickerSet(ctx context.Context, of Recipient, set *StickerSet) error
	CreateTopic(ctx context.Context, chat *Chat, topic *Topic) (*Topic, error)
	CustomEmojiStickers(ctx context.Context, ids []string) ([]Sticker, error)
	DeclineJoinRequest(ctx context.Context, chat Recipient, user *User) error
	DefaultRights(ctx context.Context, forChannels bool) (*Rights, error)
	Delete(ctx context.Context, msg Editable) error
	DeleteCommands(ctx context.Context, opts ...interface{}) error
	DeleteGroupPhoto(ctx context.Context, chat *Chat) error
	DeleteGroupStickerSet(ctx context.Context, chat *Chat) error
	DeleteMany(ctx context.Context, msgs []Editable) error
	DeleteSticker(ctx context.Context, sticker string) error
	DeleteStickerSet(ctx context.Context, name string) error
	DeleteTopic(ctx context.Context, chat *Chat, topic *Topic) error
	Download(ctx context.Context, file *File, localFilename string) error
	Edit(ctx context.Context, msg Editable, what interface{}, opts ...interface{}) (*Message, error)
	EditCaption(ctx context.Context, msg Editable, caption string, opts ...interface{}) (*Message, error)
	EditGeneralTopic(ctx context.Context, chat *Chat, topic *Topic) error
	EditInviteLink(ctx context.Context, chat Recipient, link *ChatInviteLink) (*ChatInviteLink, error)
	EditMedia(ctx context.Context, msg Editable, media Inputtable, opts ...interface{}) (*Message, error)
	EditReplyMarkup(ctx context.Context, msg Editable, markup *ReplyMarkup) (*Message, error)
	EditTopic(ctx context.Context, chat *Chat, topic *Topic) error
	File(ctx context.Context, file *File) (io.ReadCloser, error)
	FileByID(ctx context.Context, fileID string) (File, error)
	Forward(ctx context.Context, to Recipient, msg Editable, opts ...interface{}) (*Message, error)
	ForwardMany(ctx context.Context, to Recipient, msgs []Editable, opts ...*SendOptions) ([]Message, error)
	GameScores(ctx context.Context, user Recipient, msg Editable) ([]GameHighScore, error)
	HideGeneralTopic(ctx context.Context, chat *Chat) error
	InviteLink(ctx context.Context, chat *Chat) (string, error)
	Leave(ctx context.Context, chat Recipient) error
	Len(ctx context.Context, chat *Chat) (int, error)
	Logout(ctx context.Context) (bool, error)
	MenuButton(ctx context.Context, chat *User) (*MenuButton, error)
	MyDescription(ctx context.Context, language string) (*BotInfo, error)
	MyName(ctx context.Context, language string) (*BotInfo, error)
	MyShortDescription(ctx context.Context, language string) (*BotInfo, error)
	Notify(ctx context.Context, to Recipient, action ChatAction, threadID ...int) error
	Pin(ctx context.Context, msg Editable, opts ...interface{}) error
	ProfilePhotosOf(ctx context.Context, user *User) ([]Photo, error)
	Promote(ctx context.Context, chat *Chat, member *ChatMember) error
	React(ctx context.Context, to Recipient, msg Editable, r Reactions) error
	RefundStars(ctx context.Context, to Recipient, chargeID string) error
	RemoveWebhook(ctx context.Context, dropPending ...bool) error
	ReopenGeneralTopic(ctx context.Context, chat *Chat) error
	ReopenTopic(ctx context.Context, chat *Chat, topic *Topic) error
	ReplaceStickerInSet(ctx context.Context, of Recipient, stickerSet, oldSticker string, sticker InputSticker) (bool, error)
	Reply(ctx context.Context, to *Message, what interface{}, opts ...interface{}) (*Message, error)
	Respond(ctx context.Context, c *Callback, resp ...*CallbackResponse) error
	Restrict(ctx context.Context, chat *Chat, member *ChatMember) error
	RevokeInviteLink(ctx context.Context, chat Recipient, link string) (*ChatInviteLink, error)
	Send(ctx context.Context, to Recipient, what interface{}, opts ...interface{}) (*Message, error)
	SendAlbum(ctx context.Context, to Recipient, a Album, opts ...interface{}) ([]Message, error)
	SendPaid(ctx context.Context, to Recipient, stars int, a PaidAlbum, opts ...interface{}) (*Message, error)
	SetAdminTitle(ctx context.Context, chat *Chat, user *User, title string) error
	SetCommands(ctx context.Context, opts ...interface{}) error
	SetCustomEmojiStickerSetThumb(ctx context.Context, name, id string) error
	SetDefaultRights(ctx context.Context, rights Rights, forChannels bool) error
	SetGameScore(ctx context.Context, user Recipient, msg Editable, score GameHighScore) (*Message, error)
	SetGroupDescription(ctx context.Context, chat *Chat, description string) error
	SetGroupPermissions(ctx context.Context, chat *Chat, perms Rights) error
	SetGroupStickerSet(ctx context.Context, chat *Chat, setName string) error
	SetGroupTitle(ctx context.Context, chat *Chat, title string) error
	SetMenuButton(ctx context.Context, chat *User, mb interface{}) error
	SetMyDescription(ctx context.Context, desc, language string) error
	SetMyName(ctx context.Context, name, language string) error
	SetMyShortDescription(ctx context.Context, desc, language string) error
	SetStickerEmojis(ctx context.Context, sticker string, emojis []string) error
	SetStickerKeywords(ctx context.Context, sticker string, keywords []string) error
	SetStickerMaskPosition(ctx context.Context, sticker string, mask MaskPosition) error
	SetStickerPosition(ctx context.Context, sticker string, position int) error
	SetStickerSetThumb(ctx context.Context, of Recipient, set *StickerSet) error
	SetStickerSetTitle(ctx context.Context, s StickerSet) error
	SetWebhook(ctx context.Context, w *Webhook) error
	Ship(ctx context.Context, query *ShippingQuery, what ...interface{}) error
	StarTransactions(ctx context.Context, offset, limit int) ([]StarTransaction, error)
	StickerSet(ctx context.Context, name string) (*StickerSet, error)
	StopLiveLocation(ctx context.Context, msg Editable, opts ...interface{}) (*Message, error)
	StopPoll(ctx context.Context, msg Editable, opts ...interface{}) (*Poll, error)
	TopicIconStickers(ctx context.Context) ([]Sticker, error)
	Unban(ctx context.Context, chat *Chat, user *User, forBanned ...bool) error
	UnbanSenderChat(ctx context.Context, chat *Chat, sender Recipient) error
	UnhideGeneralTopic(ctx context.Context, chat *Chat) error
	Unpin(ctx context.Context, chat Recipient, messageID ...int) error
	UnpinAll(ctx context.Context, chat Recipient) error
	UnpinAllGeneralTopicMessages(ctx context.Context, chat *Chat) error
	UnpinAllTopicMessages(ctx context.Context, chat *Chat, topic *Topic) error
	UploadSticker(ctx context.Context, to Recipient, format StickerSetFormat, f File) (*File, error)
	UserBoosts(ctx context.Context, chat, user Recipient) ([]Boost, error)
	Webhook(ctx context.Context) (*Webhook, error)
}
