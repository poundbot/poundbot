package mgo

import (
	"github.com/globalsign/mgo"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

// A Chats implements db.ChatsStore
type Chats struct {
	collection *mgo.Collection // the chats collection
}

// Log implements db.ChatsStore.Log by inserting into the chats collection
func (c Chats) Log(cm types.ChatMessage) error {
	return c.collection.Insert(cm)
}
