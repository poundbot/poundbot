package rustconn

import (
	"context"

	"github.com/poundbot/poundbot/types"
)

type contextKey string

func (c contextKey) String() string {
	return "rustconn package context key " + string(c)
}

var (
	contextKeyRequestUUID = contextKey("requestUUID")
	contextKeyServerKey   = contextKey("serverKey")
	contextKeyAccount     = contextKey("account")
	contextKeyGame        = contextKey("game")
)

type serverContext struct {
	game        string
	serverKey   string
	requestUUID string
	account     types.Account
	server      types.Server
}

func getServerContext(ctx context.Context) (serverContext, error) {
	sc := serverContext{
		game:        ctx.Value(contextKeyGame).(string),
		serverKey:   ctx.Value(contextKeyServerKey).(string),
		requestUUID: ctx.Value(contextKeyRequestUUID).(string),
	}
	sc.account = ctx.Value(contextKeyAccount).(types.Account)
	var err error
	sc.server, err = sc.account.ServerFromKey(sc.serverKey)
	return sc, err
}
