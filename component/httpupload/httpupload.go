/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package httpupload

import (
	"github.com/ortuman/jackal/module/xep0030/infoprovider"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

const mailboxSize = 2048

const httpUploadServiceName = "HTTP File Upload"

const httpUploadFeature = "urn:xmpp:http:upload:0"

type uploadInfoProvider struct{}

func (ip *uploadInfoProvider) Identities(toJID, fromJID *jid.JID, node string) []infoprovider.Identity {
	return []infoprovider.Identity{
		{Category: "store", Type: "file", Name: httpUploadServiceName},
	}
}

func (ip *uploadInfoProvider) Items(toJID, fromJID *jid.JID, node string) ([]infoprovider.Item, *xmpp.StanzaError) {
	return nil, nil
}

func (ip *uploadInfoProvider) Features(toJID, fromJID *jid.JID, node string) ([]infoprovider.Feature, *xmpp.StanzaError) {
	return []infoprovider.Feature{httpUploadFeature}, nil
}

type HttpUpload struct {
	cfg          *Config
	infoProvider uploadInfoProvider
	actorCh      chan func()
	shutdownCh   <-chan struct{}
}

func New(cfg *Config, shutdownCh <-chan struct{}) *HttpUpload {
	h := &HttpUpload{
		cfg:        cfg,
		actorCh:    make(chan func(), mailboxSize),
		shutdownCh: shutdownCh,
	}
	go h.loop()
	return h
}

func (c *HttpUpload) Host() string {
	return c.cfg.Host
}

func (c *HttpUpload) ServiceName() string {
	return httpUploadServiceName
}

func (c *HttpUpload) InfoProvider() infoprovider.Provider {
	return &c.infoProvider
}

func (c *HttpUpload) ProcessStanza(stanza xmpp.Stanza) {
	c.actorCh <- func() {
		c.processStanza(stanza)
	}
}

func (c *HttpUpload) loop() {
	for {
		select {
		case f := <-c.actorCh:
			f()
		case <-c.shutdownCh:
			return
		}
	}
}

func (c *HttpUpload) processStanza(stanza xmpp.XElement) {
}
