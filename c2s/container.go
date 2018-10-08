package c2s

import (
	"sync"

	"github.com/ortuman/jackal/logger"
	"github.com/ortuman/jackal/stream"
)

var inContainer inMap

type inMap struct{ m sync.Map }

func (m *inMap) set(stm stream.C2S) {
	m.m.Store(stm.ID(), stm)
	logger.Infof("registered c2s stream... (id: %s)", stm.ID())
}

func (m *inMap) delete(stm stream.C2S) {
	m.m.Delete(stm.ID())
	logger.Infof("unregistered c2s stream... (id: %s)", stm.ID())
}
