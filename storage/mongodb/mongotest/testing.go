package mongotest

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/globalsign/mgo"
)

const defaultDial = "mongodb://localhost"
const defaultDb = "testy"

var s sync.Mutex
var dbs = [1000]bool{}

type Collection struct {
	db *mgo.Database
	C  *mgo.Collection
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
	return &Collection{db: db, C: db.C(collection), id: id}, nil
}

func (c *Collection) Close() {
	s.Lock()
	defer s.Unlock()
	dbs[c.id] = false
	c.db.DropDatabase()

}

func newDb(dbId int) (*mgo.Database, error) {
	dial := os.Getenv("MONGODB_DIAL")
	if dial == "" {
		dial = defaultDial
	}

	var sErr error
	var sess *mgo.Session
	if os.Getenv("MONGODB_SSL") != "" {
		dialInfo, err := mgo.ParseURL(dial)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
			}
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			if err != nil {
				log.Println(err)
			}
			return conn, err
		}
		sess, sErr = mgo.DialWithInfo(dialInfo)
	} else {
		sess, sErr = mgo.Dial(dial)
	}
	if sErr != nil {
		return nil, sErr
	}

	db := os.Getenv("MONGODB_DB")
	if db == "" {
		db = defaultDb
	}

	mdb := sess.DB(fmt.Sprintf("%s-%d", db, dbId))
	mdb.DropDatabase()
	return mdb, nil
}
