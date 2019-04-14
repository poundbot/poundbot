package chatcache

import (
	"fmt"
	"sync"

	"github.com/poundbot/poundbot/types"
)

type ChatCache struct {
	channelM *sync.RWMutex
	channels map[string](chan types.ChatMessage)
}

func NewChatCache() *ChatCache {
	return &ChatCache{channels: make(map[string](chan types.ChatMessage)), channelM: &sync.RWMutex{}}
}

func (c ChatCache) getChannel(name string) chan types.ChatMessage {
	c.channelM.RLock()
	schan, ok := c.channels[name]
	c.channelM.RUnlock()
	if !ok {
		schan = make(chan types.ChatMessage)
		c.channelM.Lock()
		c.channels[name] = schan
		c.channelM.Unlock()
	}
	return schan
}

func (c ChatCache) GetOutChannel(name string) chan types.ChatMessage {
	return c.getChannel(fmt.Sprintf("%s", name))
}
