/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package httpupload

import (
	"strconv"

	"github.com/ortuman/jackal/module/xep0004"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

type uploadInfoProvider struct {
	cfg *Config
}

func (ip *uploadInfoProvider) Identities(toJID, fromJID *jid.JID, node string) []xep0030.Identity {
	return []xep0030.Identity{
		{Category: "store", Type: "file", Name: httpUploadServiceName},
	}
}

func (ip *uploadInfoProvider) Items(toJID, fromJID *jid.JID, node string) ([]xep0030.Item, *xmpp.StanzaError) {
	return nil, nil
}

func (ip *uploadInfoProvider) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{httpUploadFeature}, nil
}

func (ip *uploadInfoProvider) Form(toJID, fromJID *jid.JID, node string) (*xep0004.DataForm, *xmpp.StanzaError) {
	form := &xep0004.DataForm{Type: xep0004.Result}
	fType := xep0004.Field{Var: "FORM_TYPE"}
	fType.Type = xep0004.Hidden
	fType.Values = append(fType.Values, httpUploadFeature)

	fSize := xep0004.Field{Var: "max-file-size"}
	fSize.Values = append(fSize.Values, strconv.Itoa(ip.cfg.SizeLimit))
	form.Fields = []xep0004.Field{fType, fSize}

	return form, nil
}
