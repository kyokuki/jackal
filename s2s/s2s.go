/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package s2s

import (
	"errors"
	"sync/atomic"

	"github.com/ortuman/jackal/router"

	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/module"
	"github.com/ortuman/jackal/stream"
)

const streamMailboxSize = 256

const (
	streamNamespace   = "http://etherx.jabber.org/streams"
	tlsNamespace      = "urn:ietf:params:xml:ns:xmpp-tls"
	saslNamespace     = "urn:ietf:params:xml:ns:xmpp-sasl"
	dialbackNamespace = "urn:xmpp:features:dialback"
)

type S2S struct {
	srv     *server
	enabled bool
	started uint32
}

func New(config *Config, mods *module.Modules, router *router.Router) *S2S {
	s := &S2S{}
	if config != nil {
		s.srv = &server{cfg: config, router: router, mods: mods, dialer: newDialer(config, router)}
		s.enabled = true
	}
	return s
}

func (s *S2S) Enabled() bool {
	return s.enabled
}

func (s *S2S) GetS2SOut(localDomain, remoteDomain string) (stream.S2SOut, error) {
	if s.srv == nil {
		return nil, errors.New("s2s not initialized")
	}
	return s.srv.getOrDial(localDomain, remoteDomain)
}

func (s *S2S) Start() {
	if atomic.CompareAndSwapUint32(&s.started, 0, 1) {
		go s.srv.start()
	}
}

func (s *S2S) Stop() {
	if atomic.CompareAndSwapUint32(&s.started, 1, 0) {
		if err := s.srv.stop(); err != nil {
			log.Error(err)
		}
	}
}
