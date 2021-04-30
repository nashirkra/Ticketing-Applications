package service

import (
	"log"

	"github.com/mashingan/smapping"
	"github.com/nashirkra/Ticketing-Applications/entity"
	"github.com/nashirkra/Ticketing-Applications/repository"
	"github.com/nashirkra/Ticketing-Applications/valueObjects"
)

type UserService interface {
	Insert(user valueObjects.User) (entity.User, error)
	Update(user valueObjects.User) (entity.User, error)
	Profile(userID string) entity.User
	All() []entity.User
	UserRole(userID string) string
}

type userService struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepository: userRepo,
	}
}

func (servRedis *userService) Insert(user valueObjects.User) (entity.User, error) {
	userToCreate := entity.User{}
	err := smapping.FillStruct(&userToCreate, smapping.MapFields(&user))
	if err != nil {
		log.Fatalf("Failed map %v", err)
	}
	res, err := servRedis.userRepository.InsertUser(userToCreate)
	return res, err
}

func (servRedis *userService) Update(user valueObjects.User) (entity.User, error) {
	userToUpdate := entity.User{}
	err := smapping.FillStruct(&userToUpdate, smapping.MapFields(&user))
	if err != nil {
		log.Fatalf("Failed to map %v", err)
	}
	updateUser, err := servRedis.userRepository.UpdateUser(userToUpdate)
	return updateUser, err
}
func (servRedis *userService) Profile(userID string) entity.User {
	return servRedis.userRepository.ProfileUser(userID)
}
func (servRedis *userService) All() []entity.User {
	var enusers []entity.User
	return enusers
}
func (servRedis *userService) UserRole(userID string) string {
	return ""
}
