package cache

import (
	"strings"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/bwmarrin/discordgo"
	cache "github.com/patrickmn/go-cache"
)

type Cache struct {
	userCache        *cache.Cache
	authRequestCache *cache.Cache
}

func (c Cache) Init() {
	c.userCache = cache.New(5*time.Minute, 10*time.Minute)
	c.authRequestCache = cache.New(60*time.Minute, 24*time.Hour)
}

func (c *Cache) GetUser(id string) (interface{}, bool) {
	return c.userCache.Get(id)
}
func (c *Cache) SetUser(u discordgo.User) {
	c.userCache.Set(u.String(), u, cache.DefaultExpiration)
}

func (c *Cache) GetDiscordAuth(id string) (interface{}, bool) {
	return c.authRequestCache.Get(strings.ToLower(id))
}

func (c *Cache) SetDiscordAuth(da types.DiscordAuth) {
	c.authRequestCache.Set(strings.ToLower(da.DiscordID), da, cache.DefaultExpiration)
}
