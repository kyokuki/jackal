/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package host

import (
	"crypto/tls"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/ortuman/jackal/util"
	"github.com/pkg/errors"
)

var (
	errManagerAlreadyInitialized = errors.New("host: manager already initialized")
	errManagerNotInitialized     = errors.New("host: manager not initialized")
)

const defaultDomain = "localhost"

type Manager interface {
	HostNames() []string
	IsLocalHost(domain string) bool
	Certificates() []tls.Certificate
	Close()
}

var (
	inst unsafe.Pointer
)

func Init(manager Manager) {
	if !atomic.CompareAndSwapPointer(&inst, unsafe.Pointer(nil), unsafe.Pointer(&manager)) {
		panic(errManagerAlreadyInitialized)
	}
}

func Close() {
	ptr := atomic.SwapPointer(&inst, unsafe.Pointer(nil))
	if ptr == nil {
		panic(errManagerNotInitialized)
	}
	(*(*Manager)(ptr)).Close()
}

// HostNames returns current registered domain names.
func HostNames() []string {
	return instance().HostNames()
}

// IsLocalHost returns true if domain is a local server domain.
func IsLocalHost(domain string) bool {
	return instance().IsLocalHost(domain)
}

// Certificates returns an array of all configured domain certificates.
func Certificates() []tls.Certificate {
	return instance().Certificates()
}

func instance() Manager {
	ptr := atomic.LoadPointer(&inst)
	if ptr == nil {
		panic(errManagerNotInitialized)
	}
	return *(*Manager)(ptr)
}

type manager struct {
	mu    sync.RWMutex
	hosts map[string]tls.Certificate
}

func New(configurations []Config) (Manager, error) {
	hm := &manager{
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

func (hm *manager) HostNames() []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	var ret []string
	for n, _ := range hm.hosts {
		ret = append(ret, n)
	}
	return ret
}

func (hm *manager) IsLocalHost(domain string) bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	_, ok := hm.hosts[domain]
	return ok
}

func (hm *manager) Certificates() []tls.Certificate {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	var certs []tls.Certificate
	for _, cer := range hm.hosts {
		certs = append(certs, cer)
	}
	return certs
}

func (hm *manager) Close() {}
