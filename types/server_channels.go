package types

type ServerChannel struct {
	ID       string
	Name     string
	CanSend  bool
	CanStyle bool
}

type ServerChannelsResponse struct {
	OK       bool
	Channels []ServerChannel
}

type ServerChannelsRequest struct {
	GuildID      string
	ResponseChan chan ServerChannelsResponse
}
