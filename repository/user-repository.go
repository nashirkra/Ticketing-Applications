package repository

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/nashirkra/Ticketing-Applications/entity"
	"github.com/nashirkra/Ticketing-Applications/helper"
	"github.com/nashirkra/Ticketing-Applications/valueObjects"
	"gorm.io/gorm"
)

type UserRepository interface {
	InsertUser(user entity.User) (entity.User, error)
	UpdateUser(user entity.User) (entity.User, error)
	VerifyCredential(email string, password string) interface{}
	ProfileUser(userID string) entity.User
}

type userConnection struct {
	connection *gorm.DB
	client     *redis.Client
	context    context.Context
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userConnection{
		connection: db,
		client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}),
		context: context.Background(),
	}
}

func (conn *userConnection) InsertUser(user entity.User) (entity.User, error) {
	user.Password = helper.HashAndSalt([]byte(user.Password))
	// db.connection.Save(&user)
	var userVO = valueObjects.NewUserVO(&user)
	_, err := userVO.InsertUser(conn.context, conn.client)

	return user, err
}
func (conn *userConnection) UpdateUser(user entity.User) (entity.User, error) {
	if user.Password != "" {
		user.Password = helper.HashAndSalt([]byte(user.Password))
	} else {
		var tempUser entity.User
		var userVO = valueObjects.NewUserVO(&user)
		val, err := userVO.UpdateUser(conn.context, conn.client)
		if err != nil {
			return user, err
		}
		fmt.Println(val)
		user.Password = tempUser.Password
	}

	// db.connection.Save(&user)
	return user, nil
}

func (conn *userConnection) VerifyCredential(email string, password string) interface{} {
	var user entity.User
	/*
		res := db.connection.Where("email = ?", email).Take(&user)
		if res.Error == nil {
			return user
		} */
	var userVO = valueObjects.NewUserVO(&user)
	err := userVO.FindUser(conn.context, conn.client, email)
	if err == nil {
		return user
	}
	return nil
}

func (conn *userConnection) ProfileUser(userID string) entity.User {
	var user entity.User
	var userVO = valueObjects.NewUserVO(&user)
	err := userVO.FindUser(conn.context, conn.client, "user_"+userID)
	if err == nil {
		return user
	}
	return user
}
