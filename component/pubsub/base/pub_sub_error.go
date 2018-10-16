package base

import "github.com/ortuman/jackal/xmpp"

type PubSubError struct {
	errorStanza xmpp.Stanza
}

func NewPubSubError(stanza xmpp.Stanza) *PubSubError {
	if stanza == nil {
		return nil
	}
	s := &PubSubError{
		errorStanza: stanza,
	}
	return s
}

func (e *PubSubError) ErrorStanza() xmpp.Stanza {
	return e.errorStanza
}