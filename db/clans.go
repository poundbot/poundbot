package db

import (
	"time"

	"github.com/globalsign/mgo/bson"
	"mrpoundsign.com/poundbot/types"
)

func UpsertClan(s *Session, c types.Clan) error {
	_, err := s.ClanCollection().Upsert(
		bson.M{
			"tag": c.Tag,
		},
		bson.M{
			"$setOnInsert": bson.M{"created_at": time.Now().UTC().Add(5 * time.Minute)},
			"$set":         c.BaseClan,
		},
	)
	if err != nil {
		return err
	}
	_, err = s.UserCollection().UpdateAll(
		bson.M{"steam_id": bson.M{"$in": c.Members}},
		bson.M{"$set": bson.M{"clan_tag": c.Tag}},
	)

	return err
}

func RemoveClan(s *Session, tag string) {
	s.ClanCollection().RemoveAll(bson.M{"tag": tag})
}

func RemoveClansNotIn(s *Session, tags []string) {
	s.ClanCollection().RemoveAll(
		bson.M{"tag": bson.M{"$nin": tags}},
	)
}
