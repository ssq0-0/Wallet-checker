// Package checkerUsecase implements the business logic for checking account balances
// across different blockchain services.
package checkerUsecase

import (
	"chief-checker/internal/config/usecaseConfig"
	"chief-checker/internal/domain/account"
	service "chief-checker/internal/service/checkers"
	"chief-checker/internal/usecase/features/checkerUsecase/api"
	"chief-checker/internal/usecase/features/checkerUsecase/interfaces"
	"chief-checker/internal/usecase/features/checkerUsecase/orchestration"
	"chief-checker/internal/usecase/features/checkerUsecase/processing"
	"chief-checker/internal/usecase/features/checkerUsecase/storage"
	"chief-checker/internal/usecase/selector"
	"chief-checker/pkg/errors"
	"context"
	"fmt"
	"time"
)

// Factory provides methods for creating all components needed for the checker system.
// It encapsulates the creation logic and dependencies for each component.
type Factory struct {
	config *usecaseConfig.CheckerHandlerConfig // Configuration for the checker system
}

// NewFactory creates a new instance of Factory with the provided configuration.
//
// Parameters:
// - config: configuration for the checker system
//
// Returns:
// - *Factory: initialized factory instance
func NewFactory(config *usecaseConfig.CheckerHandlerConfig) *Factory {
	return &Factory{
		config: config,
	}
}

// CreateChecker creates a specific blockchain balance checker service.
//
// Parameters:
// - name: name of the checker service to create
//
// Returns:
// - service.Checker: initialized checker service
// - error: if creation fails
func (f *Factory) CreateChecker(name string) (service.Checker, error) {
	checkerFactory := NewCheckerFactory(f.config.CheckerServiceConfig)
	return checkerFactory.CreateChecker(name)
}

// CreateErrorCollector creates a new error collector for managing errors.
//
// Returns:
// - interfaces.ErrorCollector: initialized error collector
func (f *Factory) CreateErrorCollector() interfaces.ErrorCollector {
	return storage.NewErrorCollector()
}

// CreateDataCollector creates a collector for gathering blockchain data.
//
// Parameters:
// - checker: service for checking balances
// - minUsdAmount: minimum USD amount to consider
//
// Returns:
// - interfaces.DataCollector: initialized data collector
func (f *Factory) CreateDataCollector(checker service.Checker, minUsdAmount float64) interfaces.DataCollector {
	return api.NewDataCollector(f.CreateErrorCollector(), checker, minUsdAmount)
}

// CreateAggregator creates a data aggregator for processing collected data.
//
// Parameters:
// - minUsdAmount: minimum USD amount to consider in aggregation
//
// Returns:
// - interfaces.DataAggregator: initialized data aggregator
func (f *Factory) CreateAggregator(minUsdAmount float64) interfaces.DataAggregator {
	return processing.NewDataAggregator(minUsdAmount)
}

// CreateFormatter creates a formatter for output data.
//
// Parameters:
// - minUsdAmount: minimum USD amount to include in formatting
//
// Returns:
// - interfaces.Formatter: initialized formatter
func (f *Factory) CreateFormatter(minUsdAmount float64) interfaces.Formatter {
	return processing.NewTextFormatter(minUsdAmount)
}

// CreateWriter creates a file writer for output.
//
// Parameters:
// - filename: name of the file to write to
//
// Returns:
// - interfaces.Writer: initialized writer
// - error: if writer creation fails
func (f *Factory) CreateWriter(filename string) (interfaces.Writer, error) {
	return storage.NewFileWriter(filename)
}

// CreateTaskProcessor creates a processor function for handling individual checking tasks.
// The processor coordinates data collection, aggregation, and formatting for each account.
//
// Parameters:
// - collector: collector for gathering data
// - aggregator: aggregator for processing data
// - formatter: formatter for output
//
// Returns:
// - orchestration.TaskProcessor: function for processing individual tasks
func (f *Factory) CreateTaskProcessor(
	collector interfaces.DataCollector,
	aggregator interfaces.DataAggregator,
	formatter interfaces.Formatter,
) orchestration.TaskProcessor {
	return func(ctx context.Context, acc *account.Account) ([]string, error) {
		address := acc.Address.Hex()
		accountData, err := collector.CollectData(address)
		if err != nil {
			return nil, errors.Wrap(err, "failed to collect data")
		}

		aggregatedData, err := aggregator.AggregateAccountData(address, accountData)
		if err != nil {
			return nil, errors.Wrap(err, "failed to aggregate data")
		}

		if aggregatedData == nil {
			return nil, nil
		}

		result, err := formatter.FormatAccountData(aggregatedData)
		if err != nil {
			return nil, errors.Wrap(err, "failed to format data")
		}

		return result, nil
	}
}

// CreateWorkerPool creates a task scheduler with a worker pool.
//
// Parameters:
// - processor: function for processing individual tasks
//
// Returns:
// - interfaces.TaskScheduler: initialized task scheduler
func (f *Factory) CreateWorkerPool(processor orchestration.TaskProcessor) interfaces.TaskScheduler {
	return orchestration.NewTaskScheduler(f.config.ThreadsCount, processor)
}

// CreateSystem creates and initializes the complete checker system.
// It handles user input for configuration and sets up all necessary components.
//
// Parameters:
// - config: configuration for the checker system
//
// Returns:
// - *CheckerHandler: initialized checker system
// - error: if system creation fails
func CreateSystem(config *usecaseConfig.CheckerHandlerConfig) (*CheckerHandler, error) {
	factory := NewFactory(config)

	selectedChecker, err := selector.SelectChecker("Выберите сервис для проверки балансов:")
	if err != nil {
		return nil, err
	}

	checker, err := factory.CreateChecker(selectedChecker)
	if err != nil {
		return nil, err
	}

	minUsdAmount, err := selector.SelectAmount("Введите минимальную сумму в USD для отображения:")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get min usd amount")
	}

	errorCollector := factory.CreateErrorCollector()
	collector := api.NewDataCollector(errorCollector, checker, minUsdAmount)
	aggregator := factory.CreateAggregator(minUsdAmount)
	formatter := factory.CreateFormatter(minUsdAmount)

	resultWriter, err := factory.CreateWriter(fmt.Sprintf("%s_%s.txt", "balance_checker", time.Now().Format("2006-01-02_15-04-05")))
	if err != nil {
		return nil, err
	}

	errorWriter, err := factory.CreateWriter(fmt.Sprintf("%s_%s.txt", "logs", time.Now().Format("2006-01-02_15-04-05")))
	if err != nil {
		return nil, err
	}

	processor := factory.CreateTaskProcessor(collector, aggregator, formatter)
	scheduler := factory.CreateWorkerPool(processor)
	handler, err := NewCheckerHandler(
		config.AddressFilePath,
		scheduler,
		aggregator,
		resultWriter,
		errorWriter,
		errorCollector,
	)
	if err != nil {
		return nil, err
	}

	return handler, nil
}
