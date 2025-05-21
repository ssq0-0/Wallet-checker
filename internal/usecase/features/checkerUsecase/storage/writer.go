package storage

import (
	"bufio"
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"chief-checker/pkg/errors"
	"os"
	"sync"
)

type FileWriter struct {
	filename  string        // имя выходного файла
	file      *os.File      // файловый дескриптор
	buffer    *bufio.Writer // буферизованный writer для эффективной записи
	mu        sync.Mutex    // мьютекс для синхронизации записи
	writeChan chan []string // канал для асинхронной записи
	closeChan chan struct{} // канал для сигнала о закрытии
	errChan   chan error    // канал для передачи ошибок из горутины записи
	closeOnce sync.Once     // гарантирует, что close выполняется только один раз
}

func NewFileWriter(filename string) (interfaces.Writer, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create file")
	}

	buffer := bufio.NewWriterSize(file, 64*1024) // 64KB буфер

	w := &FileWriter{
		filename:  filename,
		file:      file,
		buffer:    buffer,
		writeChan: make(chan []string, 20), // буферизованный канал для снижения блокировок
		closeChan: make(chan struct{}),
		errChan:   make(chan error, 1),
	}

	go w.writeWorker()
	return w, nil
}

func (w *FileWriter) Write(lines []string) error {
	if len(lines) == 0 {
		return nil
	}

	select {
	case err := <-w.errChan:
		return err
	default:
	}

	linesCopy := make([]string, len(lines))
	copy(linesCopy, lines)

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, line := range linesCopy {
		if _, err := w.buffer.WriteString(line + "\n"); err != nil {
			return errors.Wrap(err, "failed to write line")
		}
	}

	if err := w.buffer.Flush(); err != nil {
		return errors.Wrap(err, "failed to flush buffer")
	}

	return nil
}

func (w *FileWriter) Close() error {
	var err error

	w.closeOnce.Do(func() {
		close(w.closeChan)

		select {
		case workerErr := <-w.errChan:
			err = workerErr
		default:
		}

		if err == nil {
			w.mu.Lock()
			defer w.mu.Unlock()

			if flushErr := w.buffer.Flush(); flushErr != nil {
				err = errors.Wrap(flushErr, "failed to flush buffer")
				return
			}

			if closeErr := w.file.Close(); closeErr != nil {
				err = errors.Wrap(closeErr, "failed to close file")
			}
		}
	})

	return err
}

func (w *FileWriter) writeWorker() {
	for {
		select {
		case <-w.closeChan:
			w.mu.Lock()
			if err := w.buffer.Flush(); err != nil {
				w.errChan <- errors.Wrap(err, "failed to flush buffer on close")
			}
			w.mu.Unlock()
			return

		case lines := <-w.writeChan:
			w.mu.Lock()
			var writeErr error

			for _, line := range lines {
				if _, err := w.buffer.WriteString(line + "\n"); err != nil {
					writeErr = errors.Wrap(err, "failed to write line")
					break
				}
			}

			if writeErr == nil {
				writeErr = w.buffer.Flush()
			}

			w.mu.Unlock()

			if writeErr != nil {
				w.errChan <- writeErr
				return
			}
		}
	}
}
