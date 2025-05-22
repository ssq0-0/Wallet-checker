// Package storage provides implementations for data persistence and error handling.
package storage

import (
	"bufio"
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"chief-checker/pkg/errors"
	"os"
	"sync"
)

// FileWriter provides thread-safe, buffered file writing capabilities.
// It supports asynchronous writing through a worker goroutine and
// implements proper cleanup on close.
type FileWriter struct {
	filename  string        // Name of the output file
	file      *os.File      // File descriptor
	buffer    *bufio.Writer // Buffered writer for efficient writing
	mu        sync.Mutex    // Mutex for synchronizing write operations
	writeChan chan []string // Channel for asynchronous writing
	closeChan chan struct{} // Channel for signaling shutdown
	errChan   chan error    // Channel for propagating worker errors
	closeOnce sync.Once     // Ensures close is executed only once
}

// NewFileWriter creates a new file writer with buffered I/O.
// It initializes a background worker for asynchronous writing.
//
// Parameters:
// - filename: path to the output file
//
// Returns:
// - interfaces.Writer: initialized writer
// - error: if file creation fails
func NewFileWriter(filename string) (interfaces.Writer, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create file")
	}

	buffer := bufio.NewWriterSize(file, 64*1024) // 64KB buffer

	w := &FileWriter{
		filename:  filename,
		file:      file,
		buffer:    buffer,
		writeChan: make(chan []string, 20), // Buffered channel to reduce blocking
		closeChan: make(chan struct{}),
		errChan:   make(chan error, 1),
	}

	go w.writeWorker()
	return w, nil
}

// Write writes multiple lines to the file.
// It ensures thread safety and proper buffering of output.
//
// Parameters:
// - lines: slice of strings to write
//
// Returns:
// - error: if writing fails
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

// Close finalizes writing and releases resources.
// It ensures all buffered data is written and files are properly closed.
//
// Returns:
// - error: if closing operations fail
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

// writeWorker is a background goroutine that handles asynchronous writing.
// It processes write requests from the channel and handles shutdown signals.
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
