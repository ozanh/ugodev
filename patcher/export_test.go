package patcher

// Exported for testing purposes only.

type NumCallsGetter interface {
	NumCalls() uint64
}

func (g *goschedFunc) NumCalls() uint64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.numCalls
}
