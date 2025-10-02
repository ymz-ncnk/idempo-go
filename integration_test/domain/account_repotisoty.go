package domain

type AccountRepository interface {
	Add(account Account) error
	Get(id string) (Account, error)
	Update(account Account) error
}
