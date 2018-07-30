package db

import "mrpoundsign.com/poundbot/types"

func LogChat(s *Session, cm types.ChatMessage) {
	s.ChatCollection().Insert(cm)
}
