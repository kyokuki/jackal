/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package router

import (
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

type disabledRouter struct{}

func (_ *disabledRouter) Bind(stm stream.C2S)                             {}
func (_ *disabledRouter) Unbind(stm stream.C2S)                           {}
func (_ *disabledRouter) UserStreams(username string) []stream.C2S        { return nil }
func (_ *disabledRouter) IsBlockedJID(jid *jid.JID, username string) bool { return false }
func (_ *disabledRouter) ReloadBlockList(username string)                 {}
func (_ *disabledRouter) Route(stanza xmpp.Stanza) error                  { return nil }
func (_ *disabledRouter) MustRoute(stanza xmpp.Stanza) error              { return nil }
func (_ *disabledRouter) Close() error                                    { return nil }
