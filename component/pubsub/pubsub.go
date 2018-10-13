package pubsub

import (
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
)

const mailboxSize = 2048

const pubsubServiceName = "Publish-Subscribe"

const pubsubFeature = "http://jabber.org/protocol/pubsub"

type PubSub struct {
	cfg        *Config
	discoInfo  *xep0030.DiscoInfo
	actorCh    chan func()
	shutdownCh <-chan struct{}
}

func New(cfg *Config, discoInfo *xep0030.DiscoInfo, shutdownCh <-chan struct{}) *PubSub {
	c := &PubSub{
		cfg:        cfg,
		discoInfo:  discoInfo,
		actorCh:    make(chan func(), mailboxSize),
		shutdownCh: shutdownCh,
	}
	c.registerDiscoInfo()
	go c.loop()
	return c
}

func (c *PubSub) Host() string {
	return c.cfg.Host
}

func (c *PubSub) ProcessStanza(stanza xmpp.Stanza, stm stream.C2S) {
	c.actorCh <- func() {
		c.processStanza(stanza, stm)
	}
}

func (c *PubSub) loop() {
	for {
		select {
		case f := <-c.actorCh:
			f()
		case <-c.shutdownCh:
			c.unregisterDiscoInfo()
			return
		}
	}
}

func (c *PubSub) processStanza(stanza xmpp.XElement, stm stream.C2S) {

	switch stanza.(type) {
	case *xmpp.IQ:
		iq := stanza.(*xmpp.IQ)
		c.processIQ(iq, stm)
	case *xmpp.Message:
		msg := stanza.(*xmpp.Message)
		stm.SendElement(msg.BadRequestError())
	case *xmpp.Presence:
		prs := stanza.(*xmpp.Message)
		stm.SendElement(prs.BadRequestError())
	}
}

func (c *PubSub) processIQ(iq *xmpp.IQ, stm stream.C2S) {

	if c.discoInfo != nil && c.discoInfo.MatchesIQ(iq) {
		c.discoInfo.ProcessIQ(iq, stm)
		return
	}

	//q := iq.Elements().Child("query")
	//node := q.Attributes().Get("node")
	//if q != nil {
	//	switch q.Namespace() {
	//	case discoInfoNamespace:
	//		di.sendDiscoInfo(prov, toJID, fromJID, node, iq, stm)
	//		return
	//	case discoItemsNamespace:
	//		di.sendDiscoItems(prov, toJID, fromJID, node, iq, stm)
	//		return
	//	}
	//}
	//stm.SendElement(iq.BadRequestError())

	elem :=iq.FeatureNotImplementedError()
	stm.SendElement(elem)
}



func (c *PubSub) registerDiscoInfo() {
	c.discoInfo.RegisterServerItem(xep0030.Item{Jid: c.Host(), Name: pubsubServiceName})
	c.discoInfo.RegisterProvider(c.Host(), &pubsubInfoProvider{c.cfg})
}

func (c *PubSub) unregisterDiscoInfo() {
	c.discoInfo.UnregisterServerItem(xep0030.Item{Jid: c.Host(), Name: pubsubServiceName})
	c.discoInfo.UnregisterProvider(c.Host())
}
