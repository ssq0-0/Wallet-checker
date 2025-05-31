package selector

type Selector interface {
	SelectService(message string) (string, error)
	SelectChecker(message string) (string, error)
	SelectAmount(message string) (float64, error)
	SelectFilePath(message string) (string, error)
	SelectNumber(message string) (int, error)
	SelectServer(message string) (string, error)
}
