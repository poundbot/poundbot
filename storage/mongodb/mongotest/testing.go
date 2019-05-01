package mongotest

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultDial = "mongodb://localhost"
const defaultDb = "testy"

var s sync.Mutex
var dbs = [1000]bool{}

type Collection struct {
	db mongo.Database
	C  mongo.Collection
	id int
}

func NewCollection(collection string) (*Collection, error) {
	s.Lock()
	defer s.Unlock()

	var id int

	for i, v := range dbs {
		if !v {
			dbs[i] = true
			id = i
			break
		}
	}

	db, err := newDb(id)
	if err != nil {
		return nil, err
	}
	return &Collection{db: *db, C: *db.Collection(collection), id: id}, nil
}

func (c *Collection) Close() {
	s.Lock()
	defer s.Unlock()
	dbs[c.id] = false
	c.db.Drop(nil)

}

func newDb(dbId int) (*mongo.Database, error) {
	dial := os.Getenv("MONGODB_DIAL")
	if dial == "" {
		dial = defaultDial
	}

	db := os.Getenv("MONGODB_DB")
	if db == "" {
		db = defaultDb
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sess, err := mongo.Connect(ctx, options.Client().ApplyURI(dial))

	// sess, err := mgo.Dial(dial)
	if err != nil {
		return nil, err
	}

	mdb := sess.Database(fmt.Sprintf("%s-%d", db, dbId))
	mdb.Drop(nil)
	return mdb, nil
}
