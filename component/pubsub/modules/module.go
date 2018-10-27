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

var subModules 	map[string]AbstractModule

func AppendSubModule(modName string, modIns AbstractModule)  {
	if subModules == nil {
		subModules = make(map[string]AbstractModule)
	}
	subModules[modName] = modIns
}

func GetSubModules() map[string]AbstractModule {
	return subModules
}

func GetModuleInstance(modName string) AbstractModule {
	ins, ok := subModules[modName]
	if ok {
		return ins
	}
	return nil
}