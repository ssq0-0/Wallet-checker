package interfaces

type ErrorCollector interface {
	SaveError(context string, errorMsg string)
	WriteErrors(writer Writer) error
	HasErrors() bool
}

type Writer interface {
	Write(lines []string) error
	Close() error
}
