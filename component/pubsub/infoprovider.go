package pubsub

import (
	"github.com/ortuman/jackal/module/xep0004"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

type pubsubInfoProvider struct {
	cfg *Config
}

func (ip *pubsubInfoProvider) Identities(toJID, fromJID *jid.JID, node string) []xep0030.Identity {
	return []xep0030.Identity{
		{Category: "pubsub", Type: "service", Name: pubsubServiceName},
	}
}

func (ip *pubsubInfoProvider) Items(toJID, fromJID *jid.JID, node string) ([]xep0030.Item, *xmpp.StanzaError) {
	return nil, nil
}

func (ip *pubsubInfoProvider) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{pubsubFeature}, nil
}

func (ip *pubsubInfoProvider) Form(toJID, fromJID *jid.JID, node string) (*xep0004.DataForm, *xmpp.StanzaError) {
	form := &xep0004.DataForm{Type: xep0004.Result}
	fType := xep0004.Field{Var: "FORM_TYPE"}
	fType.Type = xep0004.Hidden
	fType.Values = append(fType.Values, pubsubFeature)

	form = nil
	return form, nil
}
