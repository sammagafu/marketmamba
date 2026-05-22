package accounts

import "forex-bot/internal/models"

// storageAccountAdapter maps generic storage with account CRUD to AccountStore.
type storageAccountAdapter struct {
	create func(*models.Account) error
	get    func(int64) (*models.Account, error)
	update func(*models.Account) error
}

func (a storageAccountAdapter) CreateAccount(account *models.Account) error {
	return a.create(account)
}

func (a storageAccountAdapter) GetAccountByUser(userID int64) (*models.Account, error) {
	return a.get(userID)
}

func (a storageAccountAdapter) UpdateAccount(account *models.Account) error {
	return a.update(account)
}

// AccountStoreFrom returns an AccountStore backed by any implementation with account CRUD.
func AccountStoreFrom(s interface{}) AccountStore {
	if s == nil {
		return nil
	}
	if a, ok := s.(AccountStore); ok {
		return a
	}
	type accountCRUD interface {
		CreateAccount(account *models.Account) error
		GetAccountByUser(userID int64) (*models.Account, error)
		UpdateAccount(account *models.Account) error
	}
	if c, ok := s.(accountCRUD); ok {
		return storageAccountAdapter{
			create: c.CreateAccount,
			get:    c.GetAccountByUser,
			update: c.UpdateAccount,
		}
	}
	return nil
}
