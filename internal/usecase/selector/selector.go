// Package selector provides interactive command-line selection utilities.
// It offers various selection methods including single-choice, multi-choice,
// and numeric input with validation.
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

// baseSelect provides a basic single-choice selection interface.
// It displays a list of options and allows the user to select one.
//
// Parameters:
// - message: prompt message to display
// - options: list of available options
//
// Returns:
// - string: selected option
// - error: if selection fails or is cancelled
func baseSelect(message string, options []string) (string, error) {
	var result string
	prompt := &survey.Select{Message: message, Options: options}

	if err := survey.AskOne(prompt, &result); err != nil {
		return "", errors.Wrap(errors.ErrSelection, err.Error())
	}
	return result, nil
}

// multiSelect provides a multiple-choice selection interface.
// It displays a list of options and allows the user to select multiple items.
//
// Parameters:
// - message: prompt message to display
// - options: list of available options
//
// Returns:
// - []string: selected options
// - error: if selection fails or no options are selected
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

// inputNumber provides a numeric input interface with validation.
// It ensures the input is a valid positive number.
//
// Parameters:
// - message: prompt message to display
//
// Returns:
// - float64: entered number
// - error: if input is invalid or not positive
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

// SelectService prompts the user to select a service from available options.
//
// Parameters:
// - message: optional custom prompt message
//
// Returns:
// - string: selected service
// - error: if selection fails
func SelectService(message string) (string, error) {
	if message == "" {
		message = "Выберите сервис:"
	}
	return baseSelect(message, append(serviceOptions, backOption))
}

// SelectChecker prompts the user to select a balance checker from available options.
//
// Parameters:
// - message: optional custom prompt message
//
// Returns:
// - string: selected checker
// - error: if selection fails
func SelectChecker(message string) (string, error) {
	if message == "" {
		message = "Выберите чекер баланса:"
	}
	return baseSelect(message, append(checkerOptions, backOption))
}

// SelectAmount prompts the user to enter a token amount.
//
// Parameters:
// - message: optional custom prompt message
//
// Returns:
// - float64: entered amount
// - error: if input is invalid
func SelectAmount(message string) (float64, error) {
	if message == "" {
		message = "Введите количество токенов:"
	}
	return inputNumber(message)
}

// SelectWaitTime prompts the user to enter a wait time in seconds.
//
// Parameters:
// - message: optional custom prompt message
//
// Returns:
// - int: wait time in seconds
// - error: if input is invalid
func SelectWaitTime(message string) (int, error) {
	if message == "" {
		message = "Введите время ожидания (в секундах):"
	}
	seconds, err := inputNumber(message)
	return int(seconds), err
}

// SelectNumber prompts the user to enter a number.
//
// Parameters:
// - message: optional custom prompt message
//
// Returns:
// - int: entered number
// - error: if input is invalid
func SelectNumber(message string) (int, error) {
	if message == "" {
		message = "Введите число:"
	}
	number, err := inputNumber(message)
	return int(number), err
}

// SelectFilePath prompts the user to enter a file path.
//
// Parameters:
// - message: optional custom prompt message
//
// Returns:
// - string: entered file path
// - error: if input fails
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
