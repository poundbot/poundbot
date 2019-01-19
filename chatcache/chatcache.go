package chatcache

import (
	"fmt"
	"sync"

	"bitbucket.org/mrpoundsign/poundbot/types"
)

type ChatCache struct {
	sync.RWMutex
	channels map[string](chan types.ChatMessage)
}

func NewChatCache() *ChatCache {
	return &ChatCache{channels: make(map[string](chan types.ChatMessage))}
}

func (c *ChatCache) getChannel(name string) chan types.ChatMessage {
	c.RLock()
	schan, ok := c.channels[name]
	c.RUnlock()
	if !ok {
		schan = make(chan types.ChatMessage)
		c.Lock()
		c.channels[name] = schan
		c.Unlock()
	}
	return schan
}

func (c *ChatCache) GetOutChannel(name string) chan types.ChatMessage {
	return c.getChannel(fmt.Sprintf("%s-out", name))
}
