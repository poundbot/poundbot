package jsonstore

import "bitbucket.org/mrpoundsign/poundbot/types"

// A Chats does nothing
type Chats struct{}

// Log does nothing
func (c Chats) Log(cm types.ChatMessage) error {
	return nil
}

func (c Chats) GetNext(serverKey string, chatMessage *types.ChatMessage) error {
	return nil
}
