package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp/jid"
	"github.com/ortuman/jackal/module/xep0030"
)

type AbstractModule interface {
	Name() string
	ModuleCriteria() *base.ElementCriteria
	Process(stanza xmpp.Stanza, stm stream.C2S) *base.PubSubError
	Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError)
}
