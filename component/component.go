/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package component

import (
	"fmt"

	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
)

// Component represents a generic component interface.
type Component interface {
	Host() string
	ProcessStanza(stanza xmpp.Stanza, stm stream.C2S)
}

type Components struct {
	comps  map[string]Component
	doneCh chan struct{}
}

func New(config *Config, discoInfo *xep0030.DiscoInfo) *Components {
	comps := &Components{
		comps:  make(map[string]Component),
		doneCh: make(chan struct{}),
	}
	cs := comps.loadComponents(config, discoInfo)
	for _, c := range cs {
		host := c.Host()
		if _, ok := comps.comps[host]; ok {
			log.Fatal(fmt.Errorf("component host name conflict: %s", host))
		}
		comps.comps[host] = c
	}
	return comps
}

// Get returns a specific component associated to host name.
func (cs *Components) Get(host string) Component {
	return cs.comps[host]
}

// GetAll returns all initialized components.
func (cs *Components) GetAll() []Component {
	var ret []Component
	for _, comp := range cs.comps {
		ret = append(ret, comp)
	}
	return ret
}

func (cs *Components) Close() {
	close(cs.doneCh)
}

func (cs *Components) loadComponents(config *Config, discoInfo *xep0030.DiscoInfo) []Component {
	var ret []Component
	/*
		discoInfo := module.Modules().DiscoInfo
		if cfg.HttpUpload != nil {
			ret = append(ret, httpupload.New(cfg.HttpUpload, discoInfo, shutdownCh))
		}
	*/
	return ret
}
