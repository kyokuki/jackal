/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package httpupload

import (
	"github.com/ortuman/jackal/xmpp"
)

const mailboxSize = 1024

const httpUploadServiceName = "HTTP File Upload"

const httpUploadFeature = "urn:xmpp:http:upload:0"

type HttpUpload struct {
	cfg        *Config
	actorCh    chan func()
	shutdownCh <-chan struct{}
}

func New(cfg *Config, shutdownCh <-chan struct{}) *HttpUpload {
	h := &HttpUpload{
		cfg:        cfg,
		actorCh:    make(chan func(), mailboxSize),
		shutdownCh: shutdownCh,
	}
	// register disco info provider
	// module.Modules().DiscoInfo.RegisterProvider(h.Host(), &infoProvider{})

	go h.loop()
	return h
}

func (c *HttpUpload) Host() string {
	return c.cfg.Host
}

func (c *HttpUpload) ServiceName() string {
	return httpUploadServiceName
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
