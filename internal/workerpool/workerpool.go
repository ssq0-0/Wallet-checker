package workerpool

import (
	"context"
	"sync"
)

type TaskFn[T any] func(ctx context.Context, item T) error

type Workerpool[T any] struct {
	tasks  chan T
	wg     sync.WaitGroup
	fn     TaskFn[T]
	mu     sync.Mutex
	closed bool
	ctx    context.Context
	cancel context.CancelFunc
}

func New[T any](workerCount, bufSize int, fn TaskFn[T]) *Workerpool[T] {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Workerpool[T]{
		tasks:  make(chan T, bufSize),
		fn:     fn,
		ctx:    ctx,
		cancel: cancel,
	}
	p.wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go p.worker()
	}

	return p
}

func (p *Workerpool[T]) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			if err := p.fn(p.ctx, task); err != nil {
				// p.cancel()
				// return
				continue
			}
		}
	}
}

func (p *Workerpool[T]) Submit(item T) error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.tasks <- item:
		return nil
		// default:
		//     return ErrQueueFull // Добавить обработку переполнения
	}
}

func (p *Workerpool[T]) Stop() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	close(p.tasks)
	p.mu.Unlock()

	p.wg.Wait()
}

func (p *Workerpool[T]) Wait() error {
	p.Stop()
	return p.ctx.Err()
}
