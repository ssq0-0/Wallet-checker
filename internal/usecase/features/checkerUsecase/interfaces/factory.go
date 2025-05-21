package interfaces

import service "chief-checker/internal/service/checkers"

type CheckerFactory interface {
	CreateChecker(name string) (service.Checker, error)
}
