package service

import (
	"log"
	"strconv"

	"github.com/mashingan/smapping"
	"github.com/nashirkra/Ticketing-Applications/entity"
	"github.com/nashirkra/Ticketing-Applications/repository"
	"github.com/nashirkra/Ticketing-Applications/valueObjects"
)

type EventService interface {
	Create(ev valueObjects.Event) (entity.Event, error)
	GetEvent(eventID string) entity.Event
	AllEvent() []entity.Event
	UserRole(userID string) string
	Profile(userID string) entity.User
}

type eventService struct {
	eventRepo repository.EventRepository
}

func NewEventService(evRepo repository.EventRepository) EventService {
	return &eventService{
		eventRepo: evRepo,
	}
}

func (serv *eventService) Create(ev valueObjects.Event) (entity.Event, error) {
	eventToCreate := entity.Event{}
	err := smapping.FillStruct(&eventToCreate, smapping.MapFields(&ev))
	if err != nil {
		log.Fatalf("Failed map %v", err)
	}

	res, err := serv.eventRepo.CreateEvent(eventToCreate)
	res.Creator = serv.eventRepo.ProfileUser(strconv.Itoa(res.CreatorId))
	return res, err
}

func (serv *eventService) GetEvent(eventID string) entity.Event {
	res := serv.eventRepo.GetEvent(eventID)
	res.Creator = serv.eventRepo.ProfileUser(strconv.Itoa(res.CreatorId))
	// fmt.Printf("(serv *eventService) GetEvent: %+v\n", res)
	if res.Creator.ID > 0 {
		return res
	}
	return res
}

func (serv *eventService) AllEvent() []entity.Event {
	return serv.eventRepo.AllEvent()
}

func (serv *eventService) UserRole(userID string) string {
	res := serv.eventRepo.ProfileUser(userID)
	return res.Role
}

func (serv *eventService) Profile(userID string) entity.User {
	return serv.eventRepo.ProfileUser(userID)
}
