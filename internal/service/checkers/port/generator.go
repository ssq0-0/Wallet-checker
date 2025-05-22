package port

import "chief-checker/internal/service/checkers/checkerModels/requestModels"

// ParamGenerator определяет интерфейс для генерации параметров запроса.
type ParamGenerator interface {
	// Generate создает параметры запроса для подписи и nonce.
	Generate(payload map[string]string, method, path string) (*requestModels.RequestParams, error)
}

// IDGenerate определяет интерфейс для генерации идентификаторов.
type IDGenerate interface {
	// Generate создает идентификатор заданной длины.
	Generate(length int) string
}
