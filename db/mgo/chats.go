package mgo

import (
	"github.com/globalsign/mgo"
	"mrpoundsign.com/poundbot/types"
)

// A Chats implements db.ChatsAccessLayer
type Chats struct {
	collection *mgo.Collection // the chats collection
}

// Log implements db.ChatsAccessLayer.Log by inserting into the chats collection
func (c Chats) Log(cm types.ChatMessage) error {
	return c.collection.Insert(cm)
}
