package proxyPool

import (
	"chief-checker/pkg/errors"
	"chief-checker/pkg/logger"
	"container/heap"
	"sync"
	"time"
)

type ProxyPool interface {
	GetFreeProxy() string
	BlockProxy(proxy string, duration time.Duration)
	UnblockProxy(proxy string)
}

type ProxyPoolImpl struct {
	mu      sync.Mutex
	minHeap *MinHeap
	proxies map[string]*Item
	cond    *sync.Cond
}

func NewProxyPool(proxyList []string) (*ProxyPoolImpl, error) {
	if len(proxyList) == 0 {
		return nil, errors.Wrap(errors.ErrInvalidParams, "empty proxy list")
	}

	minHeap := &MinHeap{}
	heap.Init(minHeap)
	proxyMap := make(map[string]*Item)

	for _, proxy := range proxyList {
		item := &Item{Proxy: proxy, Count: 0}
		heap.Push(minHeap, item)
		proxyMap[proxy] = item
	}

	pp := &ProxyPoolImpl{
		minHeap: minHeap,
		proxies: proxyMap,
	}
	pp.cond = sync.NewCond(&pp.mu)
	return pp, nil
}

func (p *ProxyPoolImpl) GetFreeProxy() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.minHeap.Len() == 0 {
		logger.GlobalLogger.Warn("Proxy pool is empty, no proxies available")
		return ""
	}

	item := heap.Pop(p.minHeap).(*Item)
	item.Count++
	heap.Push(p.minHeap, item)

	logger.GlobalLogger.Debugf("Using proxy: %s (count: %d)", item.Proxy, item.Count)
	return item.Proxy
}

func (p *ProxyPoolImpl) BlockProxy(proxy string, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	item, exists := p.proxies[proxy]
	if !exists {
		logger.GlobalLogger.Warnf("Attempted to block non-existent proxy: %s", proxy)
		return
	}

	heap.Remove(p.minHeap, item.Index)
	logger.GlobalLogger.Infof("Blocking proxy %s for %v", proxy, duration)
	delete(p.proxies, proxy)

	time.AfterFunc(duration, func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.proxies[proxy] = item
		heap.Push(p.minHeap, item)
		logger.GlobalLogger.Infof("Proxy %s is available again", proxy)
		p.cond.Signal()
	})
}

func (p *ProxyPoolImpl) UnblockProxy(proxy string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	item, exists := p.proxies[proxy]
	if !exists {
		return
	}

	if item.Count > 0 {
		item.Count--
		heap.Fix(p.minHeap, item.Index)
		p.cond.Signal()
	}
}

type Item struct {
	Proxy string
	Count int
	Index int
}

type MinHeap []*Item

func (m MinHeap) Len() int           { return len(m) }
func (m MinHeap) Less(i, j int) bool { return m[i].Count < m[j].Count }
func (m MinHeap) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
	m[i].Index = i
	m[j].Index = j
}

func (m *MinHeap) Push(x interface{}) {
	n := len(*m)
	item := x.(*Item)
	item.Index = n
	*m = append(*m, item)
}

func (m *MinHeap) Pop() interface{} {
	old := *m
	n := len(old)
	item := old[n-1]
	item.Index = -1
	*m = old[0 : n-1]
	return item
}
