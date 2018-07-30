package db

import (
	"time"

	"github.com/globalsign/mgo/bson"
	"mrpoundsign.com/poundbot/types"
)

func GetUser(s *Session, si types.SteamInfo) (*types.User, error) {
	var u types.User
	err := s.UserCollection().Find(si).One(&u)
	return &u, err
}

func BaseUserUpsert(s *Session, u types.BaseUser) error {
	_, err := s.UserCollection().Upsert(
		u.SteamInfo,
		bson.M{
			"$setOnInsert": bson.M{"created_at": time.Now().UTC().Add(5 * time.Minute)},
			"$set":         u,
		},
	)
	return err
}

func RemoveUsersClan(s *Session, tag string) {
	s.UserCollection().UpdateAll(
		bson.M{"clan_tag": tag},
		bson.M{"$unset": bson.M{"clan_tag": 1}},
	)
}

func RemoveUsersClansNotIn(s *Session, tags []string) {
	s.UserCollection().UpdateAll(
		bson.M{"clan_tag": bson.M{"$nin": tags}},
		bson.M{"$unset": bson.M{"clan_tag": 1}},
	)
}
