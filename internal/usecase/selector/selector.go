package selector

import (
	"chief-checker/pkg/errors"
	"chief-checker/pkg/utils"

	"github.com/AlecAivazis/survey/v2"
)

const (
	backOption    = "Back"
	invalidNumber = "введено некорректное число"
	positiveValue = "значение должно быть больше нуля"
)

var (
	serviceOptions = []string{"API Checker"}
	checkerOptions = []string{"debank"}
)

func baseSelect(message string, options []string) (string, error) {
	var result string
	prompt := &survey.Select{Message: message, Options: options}

	if err := survey.AskOne(prompt, &result); err != nil {
		return "", errors.Wrap(errors.ErrSelection, err.Error())
	}
	return result, nil
}

func multiSelect(message string, options []string) ([]string, error) {
	var result []string
	prompt := &survey.MultiSelect{Message: message, Options: options}

	if err := survey.AskOne(prompt, &result); err != nil {
		return nil, errors.Wrap(errors.ErrSelection, err.Error())
	}
	if len(result) == 0 {
		return nil, errors.New("необходимо выбрать хотя бы один вариант")
	}
	return result, nil
}

func inputNumber(message string) (float64, error) {
	var input string
	if err := survey.AskOne(&survey.Input{Message: message}, &input); err != nil {
		return 0, errors.Wrap(errors.ErrSelection, err.Error())
	}

	number, err := utils.ParseFloat(input)
	if err != nil {
		return 0, errors.Wrap(errors.ErrSelection, invalidNumber)
	}
	if number <= 0 {
		return 0, errors.Wrap(errors.ErrSelection, positiveValue)
	}
	return number, nil
}

func SelectService(message string) (string, error) {
	if message == "" {
		message = "Выберите сервис:"
	}
	return baseSelect(message, append(serviceOptions, backOption))
}

// SelectChecker выбирает чекер из списка доступных
func SelectChecker(message string) (string, error) {
	if message == "" {
		message = "Выберите чекер баланса:"
	}
	return baseSelect(message, append(checkerOptions, backOption))
}

func SelectAmount(message string) (float64, error) {
	if message == "" {
		message = "Введите количество токенов:"
	}
	return inputNumber(message)
}

func SelectWaitTime(message string) (int, error) {
	if message == "" {
		message = "Введите время ожидания (в секундах):"
	}
	seconds, err := inputNumber(message)
	return int(seconds), err
}

func SelectNumber(message string) (int, error) {
	if message == "" {
		message = "Введите число:"
	}
	number, err := inputNumber(message)
	return int(number), err
}

func SelectFilePath(message string) (string, error) {
	if message == "" {
		message = "Введите путь к файлу:"
	}
	var path string
	prompt := &survey.Input{Message: message}
	if err := survey.AskOne(prompt, &path); err != nil {
		return path, errors.Wrap(errors.ErrSelection, err.Error())
	}
	return path, nil
}
