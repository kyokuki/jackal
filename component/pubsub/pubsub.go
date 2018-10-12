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
}

func (c *PubSub) registerDiscoInfo() {
	c.discoInfo.RegisterServerItem(xep0030.Item{Jid: c.Host(), Name: pubsubServiceName})
	c.discoInfo.RegisterProvider(c.Host(), &pubsubInfoProvider{c.cfg})
}

func (c *PubSub) unregisterDiscoInfo() {
	c.discoInfo.UnregisterServerItem(xep0030.Item{Jid: c.Host(), Name: pubsubServiceName})
	c.discoInfo.UnregisterProvider(c.Host())
}
