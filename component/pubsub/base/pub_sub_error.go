package base

import "github.com/ortuman/jackal/xmpp"

type PubSubError struct {
	errorStanza xmpp.Stanza
}

func NewPubSubErrorStanza(stanza xmpp.Stanza, stanzaErr *xmpp.StanzaError, errorElements []xmpp.XElement) *PubSubError {
	s := &PubSubError{
		errorStanza: xmpp.NewErrorStanzaFromStanza(stanza, stanzaErr, errorElements),
	}
	return s
}

func (e *PubSubError) ErrorStanza() xmpp.Stanza {
	return e.errorStanza
}