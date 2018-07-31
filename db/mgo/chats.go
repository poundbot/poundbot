package mgo

import (
	"github.com/globalsign/mgo"
	"mrpoundsign.com/poundbot/types"
)

type Chats struct {
	collection *mgo.Collection
}

func (c Chats) Log(cm types.ChatMessage) error {
	return c.collection.Insert(cm)
}
