/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package host

import (
	"crypto/tls"
	"log"
	"sync"

	"github.com/ortuman/jackal/util"
)

type Manager interface {
	IsLocalHost(domain string) bool
	HostNames() []string
	Certificates() []tls.Certificate
}

type hostManager struct {
	mu    sync.RWMutex
	hosts map[string]tls.Certificate
}

func NewManager(configurations []Config) Manager {
	hm := &hostManager{
		hosts: make(map[string]tls.Certificate),
	}
	if len(configurations) > 0 {
		for _, h := range configurations {
			hosts[h.Name] = h.Certificate
		}
	} else {
		cer, err := util.LoadCertificate("", "", defaultDomain)
		if err != nil {
			log.Fatalf("%v", err)
		}
		hosts[defaultDomain] = cer
	}
	return hm
}

// IsLocalHost returns true if domain is a local server domain.
func (hm *hostManager) IsLocalHost(domain string) bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	_, ok := hosts[domain]
	return ok
}

// HostNames returns current registered domain names.
func (hm *hostManager) HostNames() []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	var ret []string
	for n, _ := range hosts {
		ret = append(ret, n)
	}
	return ret
}

// Certificates returns an array of all configured domain certificates.
func (hm *hostManager) Certificates() []tls.Certificate {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	var certs []tls.Certificate
	for _, cer := range hosts {
		certs = append(certs, cer)
	}
	return certs
}
