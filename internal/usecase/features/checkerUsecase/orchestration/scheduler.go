// Package orchestration provides task scheduling and execution management.
package orchestration

import (
	"chief-checker/internal/domain/account"
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"chief-checker/pkg/logger"
	"context"
	"sync"
)

// TaskProcessor defines a function type that processes a single account task
// and returns the processing results or an error.
type TaskProcessor func(context.Context, *account.Account) ([]string, error)

// TaskSchedulerImpl implements the TaskScheduler interface for managing concurrent task execution.
// It uses an adaptive worker pool to efficiently process tasks and manage system resources.
type TaskSchedulerImpl struct {
	processor    TaskProcessor       // function to process each task
	threadsCount int                 // maximum number of concurrent threads
	resultChan   chan []string       // channel for collecting results
	workerpool   *AdaptiveWorkerPool // pool of workers for task execution
}

// NewTaskScheduler creates a new task scheduler with the specified number of threads
// and processing function.
//
// Parameters:
// - threadsCount: maximum number of concurrent threads
// - processor: function to process each task
//
// Returns an implementation of the TaskScheduler interface.
func NewTaskScheduler(threadsCount int, processor TaskProcessor) interfaces.TaskScheduler {
	return &TaskSchedulerImpl{
		processor:    processor,
		threadsCount: threadsCount,
		resultChan:   make(chan []string, 10),
	}
}

// Schedule starts processing the provided accounts using the worker pool.
// It distributes tasks among workers and returns a channel for collecting results.
//
// Parameters:
// - accounts: slice of accounts to process
//
// Returns a channel that will receive processing results.
func (s *TaskSchedulerImpl) Schedule(accounts []*account.Account) <-chan []string {
	s.workerpool = NewAdaptiveWorkerPool(s.threadsCount, len(accounts), s.processAccount)

	for _, acc := range accounts {
		logger.GlobalLogger.Infof("[%s] submitting task", acc.Address.Hex())
		s.workerpool.Submit(acc)
	}

	return s.resultChan
}

// Wait blocks until all scheduled tasks are completed.
// It closes the result channel when all tasks are done.
func (s *TaskSchedulerImpl) Wait() {
	s.workerpool.Wait()
	close(s.resultChan)
}

// processAccount handles the processing of a single account.
// It executes the processor function and sends results to the result channel.
//
// Parameters:
// - ctx: context for cancellation
// - acc: account to process
//
// Returns an error if processing fails.
func (s *TaskSchedulerImpl) processAccount(ctx context.Context, acc *account.Account) error {
	result, err := s.processor(ctx, acc)
	if err != nil {
		logger.GlobalLogger.Errorf("[%s] task processing failed: %v", acc.Address.Hex(), err)
		return err
	}

	if len(result) > 0 {
		s.resultChan <- result
	}

	return nil
}

type AdaptiveWorkerPool struct {
	minWorkers  int
	maxWorkers  int
	taskQueue   chan *account.Account
	workersDone sync.WaitGroup
	workersSync sync.Mutex
	processor   func(context.Context, *account.Account) error

	activeWorkers int
	stopped       bool
}

func NewAdaptiveWorkerPool(minWorkers int, tasksCount int, processor func(context.Context, *account.Account) error) *AdaptiveWorkerPool {
	maxWorkers := minWorkers * 2
	if maxWorkers > tasksCount {
		maxWorkers = tasksCount
	}

	if minWorkers <= 0 {
		minWorkers = 1
	}

	p := &AdaptiveWorkerPool{
		minWorkers:    minWorkers,
		maxWorkers:    maxWorkers,
		taskQueue:     make(chan *account.Account, tasksCount),
		processor:     processor,
		activeWorkers: minWorkers,
	}

	p.startWorkers(minWorkers)

	return p
}

func (p *AdaptiveWorkerPool) Submit(task *account.Account) {
	p.workersSync.Lock()
	defer p.workersSync.Unlock()

	if p.stopped {
		return
	}

	queueSize := len(p.taskQueue)
	if queueSize > p.activeWorkers && p.activeWorkers < p.maxWorkers {
		workersToAdd := min(p.maxWorkers-p.activeWorkers, 5)
		p.startWorkers(workersToAdd)
		p.activeWorkers += workersToAdd
	}

	select {
	case p.taskQueue <- task:
	default:
		go func() {
			p.taskQueue <- task
		}()
	}
}

func (p *AdaptiveWorkerPool) Wait() {
	p.workersSync.Lock()
	p.stopped = true
	close(p.taskQueue)
	p.workersSync.Unlock()

	p.workersDone.Wait()
}

func (p *AdaptiveWorkerPool) startWorkers(count int) {
	for i := 0; i < count; i++ {
		p.workersDone.Add(1)
		go p.worker()
	}
}

// worker is a goroutine that processes tasks from the task queue.
// It automatically scales down when there are no tasks to process.
func (p *AdaptiveWorkerPool) worker() {
	defer p.workersDone.Done()

	ctx := context.Background()

	for task := range p.taskQueue {
		if err := p.processor(ctx, task); err != nil {
			// Error handling is delegated to the processor
		}

		p.workersSync.Lock()
		queueSize := len(p.taskQueue)
		if queueSize == 0 && p.activeWorkers > p.minWorkers {
			p.activeWorkers--
			p.workersSync.Unlock()
			return
		}
		p.workersSync.Unlock()
	}
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
