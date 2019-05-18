package types

type GameMessageType int

const (
	Plain GameMessageType = iota
	Embed
)

type GameMessage struct {
	Type          GameMessageType
	ChannelName   string
	Message       string
	Snowflake     string       `json:"-"`
	ErrorResponse chan<- error `json:"-"`
}
