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

type CheckerHandler struct {
	scheduler      interfaces.TaskScheduler
	aggregator     interfaces.DataAggregator
	resultWriter   interfaces.Writer
	errorWriter    interfaces.Writer
	errorCollector interfaces.ErrorCollector

	writeErr   error
	writeErrMu sync.Mutex
}

type DefaultCheckerFactory struct {
	config *appConfig.Checkers
}

func NewCheckerFactory(config *appConfig.Checkers) interfaces.CheckerFactory {
	return &DefaultCheckerFactory{
		config: config,
	}
}

func (f *DefaultCheckerFactory) CreateChecker(name string) (service.Checker, error) {
	return service.InitChecker(name, f.config)
}

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

	return nil
}

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
