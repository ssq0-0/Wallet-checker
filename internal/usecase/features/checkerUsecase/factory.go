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

type Factory struct {
	config *usecaseConfig.CheckerHandlerConfig
}

func NewFactory(config *usecaseConfig.CheckerHandlerConfig) *Factory {
	return &Factory{
		config: config,
	}
}

func (f *Factory) CreateChecker(name string) (service.Checker, error) {
	checkerFactory := NewCheckerFactory(f.config.CheckerServiceConfig)
	return checkerFactory.CreateChecker(name)
}

func (f *Factory) CreateErrorCollector() interfaces.ErrorCollector {
	return storage.NewErrorCollector()
}

func (f *Factory) CreateDataCollector(checker service.Checker, minUsdAmount float64) interfaces.DataCollector {
	return api.NewDataCollector(f.CreateErrorCollector(), checker, minUsdAmount)
}

func (f *Factory) CreateAggregator(minUsdAmount float64) interfaces.DataAggregator {
	return processing.NewDataAggregator(minUsdAmount)
}

func (f *Factory) CreateFormatter(minUsdAmount float64) interfaces.Formatter {
	return processing.NewTextFormatter(minUsdAmount)
}

func (f *Factory) CreateWriter(filename string) (interfaces.Writer, error) {
	return storage.NewFileWriter(filename)
}

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

func (f *Factory) CreateTaskScheduler(processor orchestration.TaskProcessor) interfaces.TaskScheduler {
	return orchestration.NewTaskScheduler(f.config.ThreadsCount, processor)
}

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

	pathToAddresses, err := selector.SelectFilePath("Введите путь к файлу с адресами для проверки:")
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
	scheduler := factory.CreateTaskScheduler(processor)
	handler, err := NewCheckerHandler(
		pathToAddresses,
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
