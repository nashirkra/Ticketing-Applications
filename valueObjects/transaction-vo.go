package valueObjects

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/nashirkra/Ticketing-Applications/entity"
	"gorm.io/gorm"
)

type Transaction struct {
	ID            int            `json:"id,string" form:"id"`
	ParticipantId int            `json:"participant_id,string" form:"participant_id"`
	CreatorId     int            `json:"creator_id,string" form:"creator_id"`
	EventId       int            `json:"event_id,string" form:"event_id"`
	Amount        float64        `json:"amount,string" form:"amount"`
	StatusPayment string         `json:"status_payment" form:"status_payment"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" form:"deleted_at" time_format:"unixNano"`
	CreatedAt     time.Time      `json:"created_at" form:"created_at" time_format:"unixNano"`
	UpdatedAt     time.Time      `json:"updated_at" form:"updated_at" time_format:"unixNano"`
}

type TransactionVO interface {
	CreateTransaction(ctx context.Context, c *redis.Client) (entity.Transaction, error)
	UpdateTransaction(ctx context.Context, c *redis.Client) (entity.Transaction, error)
	GetTransaction(ctx context.Context, c *redis.Client, key string) (entity.Transaction, error)
}

type transactionVO struct {
	transaction *entity.Transaction
	validator   *validator.Validate
}

func NewTransactionVO(trx *entity.Transaction) TransactionVO {
	v := validator.New()
	return &transactionVO{
		transaction: trx,
		validator:   v,
	}
}

func (trx *transactionVO) CreateTransaction(ctx context.Context, c *redis.Client) (entity.Transaction, error) {
	// First generate unique 1 ticket/participant/event
	evKey := "event_" + strconv.Itoa(trx.transaction.EventId)
	userKey := "user_" + strconv.Itoa(trx.transaction.ParticipantId)
	// validate Unique Key for 1 ticket/participant/event
	uk, erruk := CheckUnique(ctx, c, "uk_transaction", userKey+":"+evKey)
	if erruk == nil {
		if len(uk) > 0 {
			return *trx.transaction, fmt.Errorf("participants have purchased tickets for this event before")
		}
	}

	// get index transaction_[id]
	keyFound, err := GetIndex(ctx, c, "idx_transaction")
	if err != nil {
		keyFound = 0
	}
	keyFound++

	// generate transaction key transaction_[id]
	key := "transaction_" + strconv.Itoa(keyFound)

	trx.transaction.ID = keyFound
	trx.transaction.CreatedAt = idn()
	trx.transaction.UpdatedAt = idn()
	trx.transaction.StatusPayment = "Pending"

	// TRY to generate hash args
	args := setArgs(trx.transaction)

	// fmt.Printf("key: %+v\nargs:%+v", key, args)
	// insert All field
	_, err = c.HSet(ctx, key, args).Result()
	if err != nil {
		return *trx.transaction, fmt.Errorf("hset: %v", err)
	}

	uk_transaction := []string{key, userKey + ":" + evKey}
	b, err := json.Marshal(uk_transaction)
	if err != nil {
		return *trx.transaction, err
	}

	_ = AddUnique(ctx, c, "uk_transaction", string(b))
	_ = SetIndex(ctx, c, "idx_transaction", keyFound)

	newTransaction, err := trx.GetTransaction(ctx, c, key)
	if err != nil {
		return *trx.transaction, err
	}
	trx.transaction = &newTransaction

	return *trx.transaction, nil
}

func (trx *transactionVO) UpdateTransaction(ctx context.Context, c *redis.Client) (entity.Transaction, error) {

	// block updating without ID
	if trx.transaction.ID == 0 {
		return *trx.transaction, errs("Transaction ID not match")
	}
	// get key first
	key := "transaction_" + strconv.Itoa(trx.transaction.ID)

	// create timestamp
	trx.transaction.UpdatedAt = idn()

	var cancelled = false
	if (trx.transaction.StatusPayment == "Canceled") || (trx.transaction.StatusPayment == "Refund & Cancelled") {
		cancelled = true
	}

	// define new value to be update
	// and determine overwrite old data

	// // TRY to generate hash args
	args := setArgs(trx.transaction)

	// check event and user not null
	isZeroEvent := reflect.ValueOf(trx.transaction).Elem().FieldByName("EventId").IsZero()
	isZeroUser := reflect.ValueOf(trx.transaction).Elem().FieldByName("ParticipantId").IsZero()
	// generate unique 1 ticket/participant/event
	// to validating Unique Key for 1 ticket/participant/event
	evKey := "event_" + strconv.Itoa(trx.transaction.EventId)
	userKey := "user_" + strconv.Itoa(trx.transaction.ParticipantId)

	if !(isZeroEvent || isZeroUser) {

		// get previous data
		oldData, err := trx.GetTransaction(ctx, c, key)
		if err != nil {
			return oldData, err
		}
		oldEvKey := "event_" + strconv.Itoa(oldData.EventId)
		oldUserKey := "user_" + strconv.Itoa(oldData.ParticipantId)
		if userKey+":"+evKey != oldUserKey+":"+oldEvKey {
			uk, erruk := CheckUnique(ctx, c, "uk_transaction", userKey+":"+evKey)
			if erruk == nil {
				if len(uk) > 0 {
					return *trx.transaction, fmt.Errorf("participants have purchased tickets for this event before")
				}
			}
			uk_transaction := []string{key, userKey + ":" + evKey}
			b, err := json.Marshal(uk_transaction)
			if err != nil {
				return *trx.transaction, err
			}

			_ = AddUnique(ctx, c, "uk_transaction", string(b))

			old_uk_transaction := []string{key, oldUserKey + ":" + oldEvKey}
			old_b, err := json.Marshal(old_uk_transaction)
			if err != nil {
				return *trx.transaction, err
			}
			_ = DelUnique(ctx, c, "uk_transaction", string(old_b))
		}
	} else {
		return *trx.transaction, fmt.Errorf("participants and event not found")
	}

	// insert All field
	_, err := c.HSet(ctx, key, args).Result()
	if err != nil {
		return *trx.transaction, fmt.Errorf("hset: %v", err)
	}

	if cancelled {
		uk_transaction := []string{key, userKey + ":" + evKey}
		b, err := json.Marshal(uk_transaction)
		if err != nil {
			return *trx.transaction, err
		}
		_ = DelUnique(ctx, c, "uk_transaction", string(b))
	}

	newTransaction, err := trx.GetTransaction(ctx, c, key)
	if err != nil {
		return *trx.transaction, err
	}
	trx.transaction = &newTransaction

	return *trx.transaction, nil
}

func (trx *transactionVO) GetTransaction(ctx context.Context, c *redis.Client, key string) (entity.Transaction, error) {
	// var t1 time.Time
	// var t2 time.Time

	res, err := c.HGetAll(ctx, key).Result()
	if err != nil {
		return *trx.transaction, err
	}
	t0 := time.Unix(0, strToint64(res["DeletedAt"]))
	if !t0.IsZero() && strToint64(res["DeletedAt"]) > 0 {
		res = nil
		return *trx.transaction, errs("deleted")
	}
	// t1 = time.Unix(0, strToint64(res["CreatedAt"]))
	// t2 = time.Unix(0, strToint64(res["UpdatedAt"]))
	// id, _ := strconv.Atoi(res["ID"])

	err = fillStruct(trx.transaction, res)
	// b, err := json.Marshal(res)
	if err != nil {
		fmt.Printf("(GetTransaction error:%s)", err)
	}
	// fmt.Printf("(GetTransaction:%+v)", trx.transaction)

	// trx.transaction.CreatedAt = t1
	// trx.transaction.UpdatedAt = t2
	// trx.transaction.ID = id
	// err = json.Unmarshal(b, &trx.transaction)

	return *trx.transaction, err
}
