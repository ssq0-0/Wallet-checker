package storage

import (
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"fmt"
	"sync"
)

type memoryErrorCollector struct {
	mu     sync.Mutex
	errors map[string][]string
}

func NewErrorCollector() interfaces.ErrorCollector {
	return &memoryErrorCollector{
		errors: make(map[string][]string),
	}
}

func (c *memoryErrorCollector) SaveError(context string, errorMsg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.errors[context] = append(c.errors[context], errorMsg)
}

func (c *memoryErrorCollector) WriteErrors(writer interfaces.Writer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.errors) == 0 {
		return writer.Write([]string{"\n\n=== NO PROCESSING ERRORS FOUND ===\n"})
	}

	var lines []string
	lines = append(lines, "\n\n=== PROCESSING ERRORS ===\n")

	for context, errors := range c.errors {
		lines = append(lines, fmt.Sprintf("\n[%s] errors:", context))
		for _, errMsg := range errors {
			lines = append(lines, "  - "+errMsg)
		}
	}

	fmt.Printf("Writing %d error lines to file\n", len(lines))

	return writer.Write(lines)
}

func (c *memoryErrorCollector) HasErrors() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.errors) > 0
}
