package process

import "sync"

// Process represents a single process.
type Process struct {
	PID, PPID int
	Comm      string
	Cmdline   string
	Threads   int
}

// Walker is something that walks the /proc directory
type Walker interface {
	Walk(func(Process)) error
}

// CachingWalker is a walker than caches a copy of the output from another
// Walker, and then allows other concurrent readers to Walk that copy.
type CachingWalker struct {
	cache     []Process
	cacheLock sync.RWMutex
	source    Walker
}

// NewCachingWalker returns a new CachingWalker
func NewCachingWalker(source Walker) *CachingWalker {
	return &CachingWalker{source: source}
}

// Walk walks a cached copy of process list
func (c *CachingWalker) Walk(f func(Process)) error {
	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()

	for _, p := range c.cache {
		f(p)
	}
	return nil
}

// Update updates cached copy of process list
func (c *CachingWalker) Update() error {
	newCache := []Process{}
	err := c.source.Walk(func(p Process) {
		newCache = append(newCache, p)
	})
	if err != nil {
		return err
	}

	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()
	c.cache = newCache
	return nil
}
