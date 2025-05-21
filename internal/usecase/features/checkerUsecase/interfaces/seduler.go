package interfaces

import "chief-checker/internal/domain/account"

type TaskScheduler interface {
	Schedule(accounts []*account.Account) <-chan []string
	Wait()
}
