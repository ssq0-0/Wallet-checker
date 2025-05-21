package port

import "chief-checker/internal/service/checkers/checkerModels/requestModels"

type ParamGenerator interface {
	Generate(payload map[string]string, method, path string) (*requestModels.RequestParams, error)
}

type IDGenerate interface {
	Generate(length int) string
}
