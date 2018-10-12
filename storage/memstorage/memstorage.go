/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package memstorage

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/ortuman/jackal/model"
	"github.com/ortuman/jackal/model/rostermodel"
	"github.com/ortuman/jackal/xmpp"
)

// ErrMockedError will be returned by any Storage method
// when mocked error is activated.
var ErrMockedError = errors.New("storage mocked error")

// Storage represents an in memory storage sub system.
type Storage struct {
	mockErr             uint32
	mu                  sync.RWMutex
	users               map[string]*model.User
	rosterItems         map[string][]rostermodel.Item
	rosterVersions      map[string]rostermodel.Version
	rosterNotifications map[string][]rostermodel.Notification
	vCards              map[string]xmpp.XElement
	privateXML          map[string][]xmpp.XElement
	offlineMessages     map[string][]*xmpp.Message
	blockListItems      map[string][]model.BlockListItem
}

// New returns a new in memory storage instance.
func New() *Storage {
	return &Storage{
		users:               make(map[string]*model.User),
		rosterItems:         make(map[string][]rostermodel.Item),
		rosterVersions:      make(map[string]rostermodel.Version),
		rosterNotifications: make(map[string][]rostermodel.Notification),
		vCards:              make(map[string]xmpp.XElement),
		privateXML:          make(map[string][]xmpp.XElement),
		offlineMessages:     make(map[string][]*xmpp.Message),
		blockListItems:      make(map[string][]model.BlockListItem),
	}
}

// Close shuts down in memory storage sub system.
func (m *Storage) Close() error {
	return nil
}

// ActivateMockedError activates in memory mocked error.
func (m *Storage) ActivateMockedError() {
	atomic.StoreUint32(&m.mockErr, 1)
}

// DeactivateMockedError deactivates in memory mocked error.
func (m *Storage) DeactivateMockedError() {
	atomic.StoreUint32(&m.mockErr, 0)
}

func (m *Storage) inWriteLock(f func() error) error {
	if atomic.LoadUint32(&m.mockErr) == 1 {
		return ErrMockedError
	}
	m.mu.Lock()
	err := f()
	m.mu.Unlock()
	return err
}

func (m *Storage) inReadLock(f func() error) error {
	if atomic.LoadUint32(&m.mockErr) == 1 {
		return ErrMockedError
	}
	m.mu.RLock()
	err := f()
	m.mu.RUnlock()
	return err
}
