/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package router

import (
	"errors"
	"io"
	"sync"

	"github.com/ortuman/jackal/host"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/storage"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

var (
	// ErrNotExistingAccount will be returned by Route method
	// if destination user does not exist.
	ErrNotExistingAccount = errors.New("router: account does not exist")

	// ErrResourceNotFound will be returned by Route method
	// if destination resource does not match any of user's available resources.
	ErrResourceNotFound = errors.New("router: resource not found")

	// ErrNotAuthenticated will be returned by Route method if
	// destination user is not available at this moment.
	ErrNotAuthenticated = errors.New("router: user not authenticated")

	// ErrBlockedJID will be returned by Route method if
	// destination JID matches any of the user's blocked JID.
	ErrBlockedJID = errors.New("router: destination jid is blocked")

	// ErrFailedRemoteConnect will be returned by Route method if
	// couldn't establish a connection to the remote server.
	ErrFailedRemoteConnect = errors.New("router: failed remote connection")
)

type Router interface {
	io.Closer

	Bind(stm stream.C2S)
	Unbind(stm stream.C2S)
	UserStreams(username string) []stream.C2S
	IsBlockedJID(jid *jid.JID, username string) bool
	ReloadBlockList(username string)
	Route(stanza xmpp.Stanza) error
	MustRoute(stanza xmpp.Stanza) error
}

// Config represents router configuration.
type Config struct {

	// GetS2SOut when set, acts as an s2s outgoing stream provider.
	GetS2SOut func(localDomain, remoteDomain string) (stream.S2SOut, error)
}

var (
	instMu sync.RWMutex
	inst   Router
)

var Disabled Router = &disabledRouter{}

func init() {
	inst = Disabled
}

func Set(router Router) {
	instMu.Lock()
	inst.Close()
	inst = router
	instMu.Unlock()
}

func Unset() {
	Set(Disabled)
}

func instance() Router {
	instMu.RLock()
	r := inst
	instMu.RUnlock()
	return r
}

// Bind marks a c2s stream as binded.
// An error will be returned in case no assigned resource is found.
func Bind(stm stream.C2S) {
	instance().Bind(stm)
}

// Unbind unbinds a previously binded c2s.
// An error will be returned in case no assigned resource is found.
func Unbind(stm stream.C2S) {
	instance().Unbind(stm)
}

// UserStreams returns all streams associated to a user.
func UserStreams(username string) []stream.C2S {
	return instance().UserStreams(username)
}

// IsBlockedJID returns whether or not the passed jid matches any
// of a user's blocking list JID.
func IsBlockedJID(jid *jid.JID, username string) bool {
	return instance().IsBlockedJID(jid, username)
}

// ReloadBlockList reloads in memory block list for a given user and starts
// applying it for future stanza routing.
func ReloadBlockList(username string) {
	instance().ReloadBlockList(username)
}

// Route routes a stanza applying server rules for handling XML stanzas.
// (https://xmpp.org/rfcs/rfc3921.html#rules)
func Route(stanza xmpp.Stanza) error {
	return instance().Route(stanza)
}

// MustRoute routes a stanza applying server rules for handling XML stanzas
// ignoring blocking lists.
func MustRoute(stanza xmpp.Stanza) error {
	return instance().MustRoute(stanza)
}

type router struct {
	cfg          *Config
	mu           sync.RWMutex
	localStreams map[string][]stream.C2S
	blockListsMu sync.RWMutex
	blockLists   map[string][]*jid.JID
}

func New(config *Config) Router {
	return &router{
		cfg:          config,
		blockLists:   make(map[string][]*jid.JID),
		localStreams: make(map[string][]stream.C2S),
	}
}

func (r *router) Bind(stm stream.C2S) {
	if len(stm.Resource()) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if authenticated := r.localStreams[stm.Username()]; authenticated != nil {
		r.localStreams[stm.Username()] = append(authenticated, stm)
	} else {
		r.localStreams[stm.Username()] = []stream.C2S{stm}
	}
	log.Infof("binded c2s stream... (%s/%s)", stm.Username(), stm.Resource())
	return
}

func (r *router) Unbind(stm stream.C2S) {
	if len(stm.Resource()) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if resources := r.localStreams[stm.Username()]; resources != nil {
		res := stm.Resource()
		for i := 0; i < len(resources); i++ {
			if res == resources[i].Resource() {
				resources = append(resources[:i], resources[i+1:]...)
				break
			}
		}
		if len(resources) > 0 {
			r.localStreams[stm.Username()] = resources
		} else {
			delete(r.localStreams, stm.Username())
		}
	}
	log.Infof("unbinded c2s stream... (%s/%s)", stm.Username(), stm.Resource())
}

func (r *router) UserStreams(username string) []stream.C2S {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.localStreams[username]
}

func (r *router) IsBlockedJID(jid *jid.JID, username string) bool {
	bl := r.getBlockList(username)
	for _, blkJID := range bl {
		if r.jidMatchesBlockedJID(jid, blkJID) {
			return true
		}
	}
	return false
}

func (r *router) ReloadBlockList(username string) {
	r.blockListsMu.Lock()
	defer r.blockListsMu.Unlock()

	delete(r.blockLists, username)
	log.Infof("block list reloaded... (username: %s)", username)
}

func (r *router) Route(stanza xmpp.Stanza) error {
	return r.route(stanza, false)
}

func (r *router) MustRoute(stanza xmpp.Stanza) error {
	return r.route(stanza, true)
}

func (r *router) Close() error {
	return nil
}

func (r *router) jidMatchesBlockedJID(j, blockedJID *jid.JID) bool {
	if blockedJID.IsFullWithUser() {
		return j.Matches(blockedJID, jid.MatchesNode|jid.MatchesDomain|jid.MatchesResource)
	} else if blockedJID.IsFullWithServer() {
		return j.Matches(blockedJID, jid.MatchesDomain|jid.MatchesResource)
	} else if blockedJID.IsBare() {
		return j.Matches(blockedJID, jid.MatchesNode|jid.MatchesDomain)
	}
	return j.Matches(blockedJID, jid.MatchesDomain)
}

func (r *router) getBlockList(username string) []*jid.JID {
	r.blockListsMu.RLock()
	bl := r.blockLists[username]
	r.blockListsMu.RUnlock()
	if bl != nil {
		return bl
	}
	blItms, err := storage.FetchBlockListItems(username)
	if err != nil {
		log.Error(err)
		return nil
	}
	bl = []*jid.JID{}
	for _, blItm := range blItms {
		j, _ := jid.NewWithString(blItm.JID, true)
		bl = append(bl, j)
	}
	r.blockListsMu.Lock()
	r.blockLists[username] = bl
	r.blockListsMu.Unlock()
	return bl
}

func (r *router) route(element xmpp.Stanza, ignoreBlocking bool) error {
	toJID := element.ToJID()
	if !ignoreBlocking && !toJID.IsServer() {
		if r.IsBlockedJID(element.FromJID(), toJID.Node()) {
			return ErrBlockedJID
		}
	}
	if !host.IsLocalHost(toJID.Domain()) {
		return r.remoteRoute(element)
	}
	rcps := r.UserStreams(toJID.Node())
	if len(rcps) == 0 {
		exists, err := storage.UserExists(toJID.Node())
		if err != nil {
			return err
		}
		if exists {
			return ErrNotAuthenticated
		}
		return ErrNotExistingAccount
	}
	if toJID.IsFullWithUser() {
		for _, stm := range rcps {
			if stm.Resource() == toJID.Resource() {
				stm.SendElement(element)
				return nil
			}
		}
		return ErrResourceNotFound
	}
	switch element.(type) {
	case *xmpp.Message:
		// send to highest priority stream
		stm := rcps[0]
		var highestPriority int8
		if p := stm.Presence(); p != nil {
			highestPriority = p.Priority()
		}
		for i := 1; i < len(rcps); i++ {
			rcp := rcps[i]
			if p := rcp.Presence(); p != nil && p.Priority() > highestPriority {
				stm = rcp
				highestPriority = p.Priority()
			}
		}
		stm.SendElement(element)

	default:
		// broadcast toJID all streams
		for _, stm := range rcps {
			stm.SendElement(element)
		}
	}
	return nil
}

func (r *router) remoteRoute(elem xmpp.Stanza) error {
	if r.cfg.GetS2SOut == nil {
		return ErrFailedRemoteConnect
	}
	localDomain := elem.FromJID().Domain()
	remoteDomain := elem.ToJID().Domain()

	out, err := r.cfg.GetS2SOut(localDomain, remoteDomain)
	if err != nil {
		log.Error(err)
		return ErrFailedRemoteConnect
	}
	out.SendElement(elem)
	return nil
}
