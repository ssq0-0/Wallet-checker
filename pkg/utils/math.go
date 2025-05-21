package utils

import (
	"fmt"
	"strconv"
)

func FormatFloat(val float64, decimals int) string {
	return strconv.FormatFloat(val, 'f', decimals, 64)
}

func ParseFloat(input string) (float64, error) {
	amount, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("не удалось преобразовать '%s' в число: %w", input, err)
	}
	return amount, nil
}
