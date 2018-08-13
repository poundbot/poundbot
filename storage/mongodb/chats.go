package mongodb

import (
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// A Chats implements db.ChatsStore
type Chats struct {
	collection *mgo.Collection // the chats collection
}

// Log implements db.ChatsStore.Log by inserting into the chats collection
func (c Chats) Log(cm types.ChatMessage) error {
	return c.collection.Insert(cm)
}

func (c Chats) GetNext(serverKey string, chatMessage *types.ChatMessage) error {
	change := mgo.Change{Remove: true}
	_, err := c.collection.Find(bson.M{"serverkey": serverKey}).Limit(1).Apply(change, &chatMessage)
	return err
}