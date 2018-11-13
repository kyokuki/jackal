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

var streamC2S stream.C2S
var subModules 	map[string]AbstractModule

func InitStreamC2S(stm stream.C2S) {
	streamC2S = stm
}

func GetStreamC2S() stream.C2S {
	return streamC2S
}

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

func PacketInstance(elem xmpp.XElement, stanzaFrom jid.JID, stanzaTo jid.JID) xmpp.XElement {
	retElem, ok := elem.(*xmpp.Element)
	if ok {
		retElem.SetAttribute("from", stanzaFrom.String())
		retElem.SetAttribute("to", stanzaTo.String())
		return retElem
	}
	return elem
}