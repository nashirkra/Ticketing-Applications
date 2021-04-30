package repository

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/nashirkra/Ticketing-Applications/entity"
	"github.com/nashirkra/Ticketing-Applications/valueObjects"
)

type TransactionRepository interface {
	CreateTransaction(trx entity.Transaction) (entity.Transaction, error)
	UpdateTransaction(trx entity.Transaction) (entity.Transaction, error)
	GetTransaction(trxID string) (entity.Transaction, error)
}

type trxConnection struct {
	// connection *gorm.DB
	client  *redis.Client
	context context.Context
}

func NewTransactionRepository() TransactionRepository {
	return &trxConnection{
		client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}),
		context: context.Background(),
	}
}

func (conn *trxConnection) CreateTransaction(trx entity.Transaction) (entity.Transaction, error) {
	var trxVO = valueObjects.NewTransactionVO(&trx)

	_, err := trxVO.CreateTransaction(conn.context, conn.client)
	return trx, err
}
func (conn *trxConnection) UpdateTransaction(trx entity.Transaction) (entity.Transaction, error) {
	var trxVO = valueObjects.NewTransactionVO(&trx)

	_, err := trxVO.UpdateTransaction(conn.context, conn.client)
	return trx, err
}
func (conn *trxConnection) GetTransaction(trxID string) (entity.Transaction, error) {
	var trx entity.Transaction
	var trxVO = valueObjects.NewTransactionVO(&trx)

	trx, err := trxVO.GetTransaction(conn.context, conn.client, "transaction_"+trxID)
	return trx, err
}
