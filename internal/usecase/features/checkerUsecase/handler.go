// Package checkerUsecase implements the business logic for checking account balances
// across different blockchain services.
package checkerUsecase

import (
	"chief-checker/internal/config/appConfig"
	"chief-checker/internal/domain/account"
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"chief-checker/internal/usecase/features/checkerUsecase/processing"
	"chief-checker/internal/usecase/features/checkerUsecase/types"

	service "chief-checker/internal/service/checkers"
	"chief-checker/pkg/errors"
	"chief-checker/pkg/logger"
	"chief-checker/pkg/utils"
	"sync"
)

// CheckerHandler coordinates the checking of account balances and processing of results.
// It manages the scheduling of tasks, aggregation of data, and writing of results and errors.
type CheckerHandler struct {
	scheduler      interfaces.TaskScheduler  // Schedules and manages checking tasks
	aggregator     interfaces.DataAggregator // Aggregates checking results
	resultWriter   interfaces.Writer         // Writes successful results
	errorWriter    interfaces.Writer         // Writes errors
	errorCollector interfaces.ErrorCollector // Collects and manages errors

	writeErr   error      // Stores any write errors that occur
	writeErrMu sync.Mutex // Protects writeErr from concurrent access
}

// DefaultCheckerFactory creates checker services based on configuration.
type DefaultCheckerFactory struct {
	config *appConfig.Checkers // Checker service configuration
}

// NewCheckerFactory creates a new instance of DefaultCheckerFactory.
//
// Parameters:
// - config: configuration for checker services
//
// Returns:
// - interfaces.CheckerFactory: factory instance for creating checkers
func NewCheckerFactory(config *appConfig.Checkers) interfaces.CheckerFactory {
	return &DefaultCheckerFactory{
		config: config,
	}
}

// CreateChecker creates a specific checker service by name.
//
// Parameters:
// - name: name of the checker service to create
//
// Returns:
// - service.Checker: initialized checker service
// - error: if creation fails
func (f *DefaultCheckerFactory) CreateChecker(name string) (service.Checker, error) {
	return service.InitChecker(name, f.config)
}

// NewCheckerHandler creates a new instance of CheckerHandler.
// It initializes the handler with accounts from the provided file and starts processing results.
//
// Parameters:
// - addressesPath: path to file containing addresses to check
// - scheduler: task scheduler for managing checking operations
// - aggregator: data aggregator for collecting results
// - resultWriter: writer for successful results
// - errorWriter: writer for errors
// - errorCollector: collector for error management
//
// Returns:
// - *CheckerHandler: initialized handler
// - error: if initialization fails
func NewCheckerHandler(
	addressesPath string,
	scheduler interfaces.TaskScheduler,
	aggregator interfaces.DataAggregator,
	resultWriter interfaces.Writer,
	errorWriter interfaces.Writer,
	errorCollector interfaces.ErrorCollector,
) (*CheckerHandler, error) {
	addresses, err := utils.FileReader(addressesPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read addresses file")
	}

	accounts, err := account.AccountFactory(addresses, account.AccountWithAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create accounts")
	}

	resultChan := scheduler.Schedule(accounts)

	handler := &CheckerHandler{
		scheduler:      scheduler,
		aggregator:     aggregator,
		resultWriter:   resultWriter,
		errorWriter:    errorWriter,
		errorCollector: errorCollector,
	}

	go handler.processResults(resultChan)

	return handler, nil
}

// Handle executes the main checking workflow.
// It waits for all tasks to complete and writes final results and errors.
//
// Returns:
// - error: if any operation fails during execution
func (h *CheckerHandler) Handle() error {
	h.scheduler.Wait()

	h.writeErrMu.Lock()
	defer h.writeErrMu.Unlock()
	if h.writeErr != nil {
		return h.writeErr
	}

	if err := h.writeGlobalStats(); err != nil {
		return err
	}

	hasErrors := h.errorCollector.HasErrors()
	logger.GlobalLogger.Debugf("Has errors to write: %v", hasErrors)

	if hasErrors {
		logger.GlobalLogger.Debugf("Writing errors to file...")
		if err := h.errorCollector.WriteErrors(h.errorWriter); err != nil {
			return err
		}
		if err := h.errorWriter.Close(); err != nil {
			return err
		}
		logger.GlobalLogger.Debugf("Errors written successfully")
	} else {
		if err := h.errorWriter.Close(); err != nil {
			return err
		}
	}

	if err := h.resultWriter.Close(); err != nil {
		return err
	}
	logger.GlobalLogger.Infof("Global stats: %+v", h.aggregator.GetGlobalStats())
	return nil
}

// writeGlobalStats writes aggregated statistics to the result writer.
//
// Returns:
// - error: if writing fails
func (h *CheckerHandler) writeGlobalStats() error {
	globalStats := h.aggregator.GetGlobalStats()

	formatter := processing.NewTextFormatter(0)
	formattedStats, err := formatter.FormatGlobalStats(globalStats)
	if err != nil {
		return err
	}

	if err := h.resultWriter.Write(formattedStats); err != nil {
		return err
	}

	return nil
}

// processResults processes checking results in batches.
// It reads results from the channel and writes them in batches for efficiency.
//
// Parameters:
// - results: channel providing checking results
func (h *CheckerHandler) processResults(results <-chan []string) {
	var batch [][]string
	batchSize := types.BatchSize

	for result := range results {
		batch = append(batch, result)
		if len(batch) >= batchSize {
			h.writeBatch(batch)
			batch = nil
		}
	}

	if len(batch) > 0 {
		h.writeBatch(batch)
	}
}

// writeBatch writes a batch of results to the result writer.
// If a write error occurs, it is stored for later handling.
//
// Parameters:
// - batch: slice of result slices to write
func (h *CheckerHandler) writeBatch(batch [][]string) {
	var lines []string
	for _, res := range batch {
		lines = append(lines, res...)
	}

	if err := h.resultWriter.Write(lines); err != nil {
		h.writeErrMu.Lock()
		if h.writeErr == nil {
			h.writeErr = err
		}
		h.writeErrMu.Unlock()
	}
}
