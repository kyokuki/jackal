package pubsub

import (
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/component/pubsub/modules"
	"fmt"
)

const mailboxSize = 2048

const pubsubServiceName = "Publish-Subscribe"

const pubsubFeature = "http://jabber.org/protocol/pubsub"

type PubSub struct {
	cfg        *Config
	discoInfo  *xep0030.DiscoInfo
	actorCh    chan func()
	shutdownCh <-chan struct{}

	modules 	[]modules.AbstractModule
}

func New(cfg *Config, discoInfo *xep0030.DiscoInfo, shutdownCh <-chan struct{}) *PubSub {
	c := &PubSub{
		cfg:        cfg,
		discoInfo:  discoInfo,
		actorCh:    make(chan func(), mailboxSize),
		shutdownCh: shutdownCh,
	}
	c. initModules()
	c.registerDiscoInfo()
	go c.loop()
	return c
}

func (c *PubSub)initModules()  {
	c.modules = append(c.modules, &modules.DiscoveryModule{})
	c.modules = append(c.modules, &modules.NodeCreateModule{})
	c.modules = append(c.modules, &modules.NodeConfigModule{})
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

func (c *PubSub) processStanza(stanza xmpp.Stanza, stm stream.C2S) {

	var hanlde bool
	hanlde = c.process(stanza, stm)
	if !hanlde {
		switch stanza.(type) {
		case *xmpp.IQ:
			iq := stanza.(*xmpp.IQ)
			stm.SendElement(iq.FeatureNotImplementedError())
		case *xmpp.Message:
			msg := stanza.(*xmpp.Message)
			stm.SendElement(msg.FeatureNotImplementedError())
		case *xmpp.Presence:
			prs := stanza.(*xmpp.Message)
			stm.SendElement(prs.FeatureNotImplementedError())
		}
	}
}

func (c *PubSub) process(stanza xmpp.Stanza, stm stream.C2S) bool {
	handled := false
	for _, mod := range c.modules {
		criteria := mod.ModuleCriteria()
		if criteria != nil && criteria.Matches(stanza) {
			handled = true
			fmt.Println("Handled by module " + mod.Name())
			pubSubErr := mod.Process(stanza, stm)
			if pubSubErr != nil {
				stm.SendElement(pubSubErr.ErrorStanza())
			}
			fmt.Println("Finished " + mod.Name())
		}
	}
	return handled
}



func (c *PubSub) registerDiscoInfo() {
	c.discoInfo.RegisterServerItem(xep0030.Item{Jid: c.Host(), Name: pubsubServiceName})
	c.discoInfo.RegisterProvider(c.Host(), &pubsubInfoProvider{c.cfg})
}

func (c *PubSub) unregisterDiscoInfo() {
	c.discoInfo.UnregisterServerItem(xep0030.Item{Jid: c.Host(), Name: pubsubServiceName})
	c.discoInfo.UnregisterProvider(c.Host())
}
