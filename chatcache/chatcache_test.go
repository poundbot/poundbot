package chatcache

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/poundbot/poundbot/types"
	"github.com/stretchr/testify/assert"
)

func TestNewChatCache(t *testing.T) {
	tests := []struct {
		name string
		want *ChatCache
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewChatCache(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewChatCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChatCache_GetOutChannel(t *testing.T) {
	t.Parallel()
	sameChan := make(chan types.ChatMessage)

	type args struct {
		name string
	}
	tests := []struct {
		name       string
		channels   map[string](chan types.ChatMessage)
		args       args
		wantLength int
	}{
		{
			name:       "first cache channel",
			channels:   make(map[string](chan types.ChatMessage)),
			args:       args{name: "foo"},
			wantLength: 1,
		},
		{
			name:       "second cache channel",
			channels:   map[string](chan types.ChatMessage){"bar": nil},
			args:       args{name: "foo"},
			wantLength: 2,
		},
		{
			name:       "existing channel",
			channels:   map[string](chan types.ChatMessage){"foo": sameChan},
			args:       args{name: "foo"},
			wantLength: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ChatCache{
				channelM: &sync.RWMutex{},
				channels: tt.channels,
			}

			d := *c

			assert.NotEqual(t, fmt.Sprintf("%p", d), fmt.Sprintf("%p", c), "c and d are equal")

			d.GetOutChannel(tt.args.name)
			assert.Equal(t, tt.wantLength, len(c.channels), "Wrong number of channels for c")
			assert.Equal(t, tt.wantLength, len(d.channels), "Wrong number of channels for d")
		})
	}
}
