package account

type AccountDomainType string

const (
	AccountWithPrivateKey AccountDomainType = "With Private Key"
	AccountWithAddress    AccountDomainType = "With Address"
)
