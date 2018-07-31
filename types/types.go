package types

import (
	"github.com/globalsign/mgo/bson"
)

// MongoID is the _id field for mongodb
type MongoID struct {
	ID bson.ObjectId `bson:"_id,omitempty"`
}

// RESTError Generic REST error response
type RESTError struct {
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
}
