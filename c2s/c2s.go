/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package c2s

import (
	"sync"
	"sync/atomic"

	"github.com/ortuman/jackal/component"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/module"
	"github.com/ortuman/jackal/router"
	"github.com/pkg/errors"
)

const (
	streamMailboxSize   = 64
	iqResultMailboxSize = 32
)

const (
	streamNamespace           = "http://etherx.jabber.org/streams"
	tlsNamespace              = "urn:ietf:params:xml:ns:xmpp-tls"
	compressProtocolNamespace = "http://jabber.org/protocol/compress"
	bindNamespace             = "urn:ietf:params:xml:ns:xmpp-bind"
	sessionNamespace          = "urn:ietf:params:xml:ns:xmpp-session"
	saslNamespace             = "urn:ietf:params:xml:ns:xmpp-sasl"
	blockedErrorNamespace     = "urn:xmpp:blocking:errors"
)

type C2S struct {
	mu      sync.RWMutex
	servers map[string]*server
	started uint32
}

func New(configs []Config, mods *module.Modules, comps *component.Components, router *router.Router) (*C2S, error) {
	if len(configs) == 0 {
		return nil, errors.New("at least one c2s configuration is required")
	}
	c := &C2S{servers: make(map[string]*server)}
	for _, config := range configs {
		srv := &server{cfg: &config, mods: mods, comps: comps, router: router}
		c.servers[config.ID] = srv
	}
	return c, nil
}

// Start initializes c2s sub system spawning a connection listener for every server configuration.
func (c *C2S) Start() {
	if atomic.CompareAndSwapUint32(&c.started, 0, 1) {
		for _, srv := range c.servers {
			go srv.start()
		}
	}
}

// Stop closes every server listener.
func (c *C2S) Stop() {
	if atomic.CompareAndSwapUint32(&c.started, 1, 0) {
		for _, srv := range c.servers {
			if err := srv.stop(); err != nil {
				log.Error(err)
			}
		}
	}
}
