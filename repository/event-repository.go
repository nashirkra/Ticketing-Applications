package repository

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/nashirkra/Ticketing-Applications/entity"
	"github.com/nashirkra/Ticketing-Applications/valueObjects"
)

type EventRepository interface {
	CreateEvent(event entity.Event) (entity.Event, error)
	GetEvent(eventID string) entity.Event
	AllEvent() []entity.Event
	ProfileUser(userID string) entity.User
}

type eventConnection struct {
	// connection *gorm.DB
	client  *redis.Client
	context context.Context
}

func NewEventRepository() EventRepository {
	return &eventConnection{
		client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}),
		context: context.Background(),
	}
}

func (conn *eventConnection) CreateEvent(event entity.Event) (entity.Event, error) {
	var eventVO = valueObjects.NewEventVO(&event)

	if len(event.LinkWebinar) == 0 {
		//generate urls from title
		reg, err := regexp.Compile("[^A-Za-z0-9]+")
		if err != nil {
			log.Fatal(err)
		}
		genUrl := reg.ReplaceAllString(event.TitleEvent, "-")
		genUrl = strings.ToLower(strings.Trim(genUrl, "-"))
		event.LinkWebinar = "https://get.event.id/event/" + genUrl
	}

	_, err := eventVO.CreateEvent(conn.context, conn.client)
	return event, err
}
func (conn *eventConnection) GetEvent(eventID string) entity.Event {
	var event entity.Event
	var eventVO = valueObjects.NewEventVO(&event)
	event, _ = eventVO.GetEvent(conn.context, conn.client, "event_"+eventID)
	return event
}
func (conn *eventConnection) AllEvent() []entity.Event {
	var event entity.Event
	var events []entity.Event
	var eventVO = valueObjects.NewEventVO(&event)
	events, _ = eventVO.GetAllEvent(conn.context, conn.client)
	return events
}

func (conn *eventConnection) ProfileUser(userID string) entity.User {
	var user entity.User
	var userVO = valueObjects.NewUserVO(&user)
	err := userVO.FindUser(conn.context, conn.client, "user_"+userID)
	// fmt.Printf("(conn *eventConnection) ProfileUser %+v\nerror: %+v", user, err)
	if err == nil {
		return user
	}
	return user
}
