package usecase

type UseCaseInterface interface {
	Run() error
}

type HandlerInterface interface {
	Handle() error
}
