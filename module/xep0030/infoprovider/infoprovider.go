package infoprovider

import (
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

// Feature represents a disco info feature entity.
type Feature = string

// Identity represents a disco info identity entity.
type Identity struct {
	Category string
	Type     string
	Name     string
}

// Item represents a disco info item entity.
type Item struct {
	Jid  string
	Name string
	Node string
}

// Provider represents a generic disco info domain provider.
type Provider interface {
	// Identities returns all identities associated to the provider.
	Identities(toJID, fromJID *jid.JID, node string) []Identity

	// Items returns all items associated to the provider.
	// A proper stanza error should be returned in case an error occurs.
	Items(toJID, fromJID *jid.JID, node string) ([]Item, *xmpp.StanzaError)

	// Features returns all features associated to the provider.
	// A proper stanza error should be returned in case an error occurs.
	Features(toJID, fromJID *jid.JID, node string) ([]Feature, *xmpp.StanzaError)
}
