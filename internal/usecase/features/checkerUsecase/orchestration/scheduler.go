package orchestration

import (
	"chief-checker/internal/domain/account"
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"chief-checker/pkg/logger"
	"context"
	"sync"
)

// TaskProcessor обрабатывает задачу для конкретного аккаунта и возвращает результат
type TaskProcessor func(context.Context, *account.Account) ([]string, error)

// TaskScheduler реализует интерфейс types.TaskScheduler
type TaskSchedulerImpl struct {
	processor    TaskProcessor
	threadsCount int
	resultChan   chan []string
	workerpool   *AdaptiveWorkerPool
}

// NewTaskScheduler создает новый экземпляр планировщика задач
func NewTaskScheduler(threadsCount int, processor TaskProcessor) interfaces.TaskScheduler {
	return &TaskSchedulerImpl{
		processor:    processor,
		threadsCount: threadsCount,
		resultChan:   make(chan []string, 10),
	}
}

// Schedule планирует задачи для выполнения и возвращает канал с результатами
func (s *TaskSchedulerImpl) Schedule(accounts []*account.Account) <-chan []string {
	s.workerpool = NewAdaptiveWorkerPool(s.threadsCount, len(accounts), s.processAccount)

	for _, acc := range accounts {
		logger.GlobalLogger.Infof("[%s] submitting task", acc.Address.Hex())
		s.workerpool.Submit(acc)
	}

	return s.resultChan
}

// Wait ожидает завершения всех задач
func (s *TaskSchedulerImpl) Wait() {
	s.workerpool.Wait()
	close(s.resultChan)
}

// processAccount обрабатывает задачу для конкретного аккаунта
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

// AdaptiveWorkerPool - воркер-пул с адаптивной емкостью
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

// NewAdaptiveWorkerPool создает новый пул с адаптивным количеством рабочих
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

	// Запускаем минимальное количество рабочих
	p.startWorkers(minWorkers)

	return p
}

// Submit добавляет задачу в очередь на выполнение
func (p *AdaptiveWorkerPool) Submit(task *account.Account) {
	p.workersSync.Lock()
	defer p.workersSync.Unlock()

	if p.stopped {
		return
	}

	// Проверяем, нужно ли добавить рабочих
	queueSize := len(p.taskQueue)
	if queueSize > p.activeWorkers && p.activeWorkers < p.maxWorkers {
		// Добавляем рабочих, если очередь растет
		workersToAdd := min(p.maxWorkers-p.activeWorkers, 5)
		p.startWorkers(workersToAdd)
		p.activeWorkers += workersToAdd
	}

	// Изменяем на неблокирующую запись в канал
	select {
	case p.taskQueue <- task:
		// Задача успешно добавлена
	default:
		// Канал заполнен, запускаем новую горутину для добавления задачи
		go func() {
			p.taskQueue <- task
		}()
	}
}

// Wait ожидает завершения всех задач
func (p *AdaptiveWorkerPool) Wait() {
	// Закрываем канал задач, чтобы рабочие завершились
	p.workersSync.Lock()
	p.stopped = true
	close(p.taskQueue)
	p.workersSync.Unlock()

	// Ждем завершения всех рабочих
	p.workersDone.Wait()
}

// startWorkers запускает указанное количество рабочих
func (p *AdaptiveWorkerPool) startWorkers(count int) {
	for i := 0; i < count; i++ {
		p.workersDone.Add(1)
		go p.worker()
	}
}

// worker обрабатывает задачи из очереди
func (p *AdaptiveWorkerPool) worker() {
	defer p.workersDone.Done()

	ctx := context.Background()

	for task := range p.taskQueue {
		// Обрабатываем задачу
		if err := p.processor(ctx, task); err != nil {
			// logger.GlobalLogger.Errorf("Task processing error: %v", err)
		}

		// Проверяем, нужно ли уменьшить количество рабочих
		p.workersSync.Lock()
		queueSize := len(p.taskQueue)
		if queueSize == 0 && p.activeWorkers > p.minWorkers {
			// Если рабочий заметил, что очередь пуста и рабочих слишком много,
			// он завершает свою работу
			p.activeWorkers--
			p.workersSync.Unlock()
			return
		}
		p.workersSync.Unlock()
	}
}

// Вспомогательная функция min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
