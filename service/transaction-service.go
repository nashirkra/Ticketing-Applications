package service

import (
	"log"
	"strconv"

	"github.com/mashingan/smapping"
	"github.com/nashirkra/Ticketing-Applications/entity"
	"github.com/nashirkra/Ticketing-Applications/repository"
	"github.com/nashirkra/Ticketing-Applications/valueObjects"
)

type TransactionService interface {
	Create(trx valueObjects.Transaction) (entity.Transaction, error)
	Update(trx valueObjects.Transaction) (entity.Transaction, error)
	GetTicket(trxID string) (entity.Transaction, error)
}

type transactionService struct {
	trxRepo        repository.TransactionRepository
	userRepository repository.UserRepository
	eventRepo      repository.EventRepository
}

func NewTransactionService(trxRepo repository.TransactionRepository, userRepo repository.UserRepository, evRepo repository.EventRepository) TransactionService {
	return &transactionService{
		trxRepo:        trxRepo,
		userRepository: userRepo,
		eventRepo:      evRepo,
	}
}
func (serv *transactionService) Create(trx valueObjects.Transaction) (entity.Transaction, error) {
	trxToCreate := entity.Transaction{}
	err := smapping.FillStruct(&trxToCreate, smapping.MapFields(&trx))
	if err != nil {
		log.Fatalf("Failed map %v", err)
	}

	res, err := serv.trxRepo.CreateTransaction(trxToCreate)
	res.Participant = serv.userRepository.ProfileUser(strconv.Itoa(res.ParticipantId))
	res.Event = serv.eventRepo.GetEvent(strconv.Itoa(res.EventId))
	res.Event.Creator = serv.userRepository.ProfileUser(strconv.Itoa(res.Event.CreatorId))
	return res, err
}
func (serv *transactionService) Update(trx valueObjects.Transaction) (entity.Transaction, error) {
	trxToUpdate := entity.Transaction{}
	err := smapping.FillStruct(&trxToUpdate, smapping.MapFields(&trx))
	if err != nil {
		log.Fatalf("Failed map %v", err)
	}

	res, err := serv.trxRepo.UpdateTransaction(trxToUpdate)
	// res.Participant = serv.userRepository.ProfileUser(strconv.Itoa(res.ParticipantId))
	// res.Event = serv.eventRepo.GetEvent(strconv.Itoa(res.EventId))
	// res.Event.Creator = serv.userRepository.ProfileUser(strconv.Itoa(res.Event.CreatorId))
	return res, err
}
func (serv *transactionService) GetTicket(trxID string) (entity.Transaction, error) {

	res, err := serv.trxRepo.GetTransaction(trxID)
	return res, err
}
