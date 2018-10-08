/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package hostmanager

import (
	"crypto/tls"
	"sync/atomic"
	"unsafe"

	"github.com/ortuman/jackal/util"
)

const defaultDomain = "localhost"

type Manager interface {
	HostNames() []string
	IsLocalHost(domain string) bool
	Certificates() []tls.Certificate
}

var (
	instance unsafe.Pointer
)

type dummyManager struct{}

func (_ *dummyManager) HostNames() []string             { return nil }
func (_ *dummyManager) IsLocalHost(domain string) bool  { return false }
func (_ *dummyManager) Certificates() []tls.Certificate { return nil }

func init() {
	Set(&dummyManager{})
}

func Set(manager Manager) {
	atomic.StorePointer(&instance, unsafe.Pointer(&manager))
}

func get() Manager {
	return *(*Manager)(atomic.LoadPointer(&instance))
}

// HostNames returns current registered domain names.
func HostNames() []string { return get().HostNames() }

// IsLocalHost returns true if domain is a local server domain.
func IsLocalHost(domain string) bool { return get().IsLocalHost(domain) }

// Certificates returns an array of all configured domain certificates.
func Certificates() []tls.Certificate { return get().Certificates() }

type hostManager struct {
	hosts map[string]tls.Certificate
}

func New(configurations []Config) (Manager, error) {
	hm := &hostManager{
		hosts: make(map[string]tls.Certificate),
	}
	if len(configurations) > 0 {
		for _, h := range configurations {
			hm.hosts[h.Name] = h.Certificate
		}
	} else {
		cer, err := util.LoadCertificate("", "", defaultDomain)
		if err != nil {
			return nil, err
		}
		hm.hosts[defaultDomain] = cer
	}
	return hm, nil
}

func (hm *hostManager) HostNames() []string {
	var ret []string
	for n, _ := range hm.hosts {
		ret = append(ret, n)
	}
	return ret
}

func (hm *hostManager) IsLocalHost(domain string) bool {
	_, ok := hm.hosts[domain]
	return ok
}

func (hm *hostManager) Certificates() []tls.Certificate {
	var certs []tls.Certificate
	for _, cer := range hm.hosts {
		certs = append(certs, cer)
	}
	return certs
}
