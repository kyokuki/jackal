package modules

import (
	"github.com/ortuman/jackal/component/pubsub/base"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/stream"
)

type AbstractModule interface {
	Name() string
	ModuleCriteria() *base.ElementCriteria
	Process(stanza xmpp.XElement, stm stream.C2S)
}
