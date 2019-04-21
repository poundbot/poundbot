package rustconn

type contextKey string

func (c contextKey) String() string {
	return "myrustconn package context key " + string(c)
}

var (
	contextKeyRequestUUID = contextKey("requestUUID")
	contextKeyServerKey   = contextKey("serverKey")
	contextKeyAccount     = contextKey("account")
	contextKeyGame        = contextKey("game")
)
