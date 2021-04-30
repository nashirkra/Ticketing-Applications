package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/nashirkra/Ticketing-Applications/entity"
	"github.com/nashirkra/Ticketing-Applications/helper"
	"github.com/nashirkra/Ticketing-Applications/service"
	"github.com/nashirkra/Ticketing-Applications/valueObjects"
)

type UserController interface {
	All(context *gin.Context)
	Update(context *gin.Context)
	Profile(context *gin.Context)
}

type userController struct {
	userService service.UserService
	jwtService  service.JWTService
}

func NewUserController(userServ service.UserService, jwtServ service.JWTService) UserController {
	return &userController{
		userService: userServ,
		jwtService:  jwtServ,
	}
}

func (c *userController) Update(context *gin.Context) {
	var userUpdate valueObjects.User
	err := context.ShouldBind(&userUpdate)
	if err != nil {
		res := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		context.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	authHeader := context.GetHeader("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)

	if errToken != nil {
		panic(errToken.Error())
	}
	claims := token.Claims.(jwt.MapClaims)
	id, err := strconv.Atoi(fmt.Sprintf("%v", claims["user_id"]))
	if err != nil {
		panic(err.Error())
	}
	userUpdate.ID = id
	u, err := c.userService.Update(userUpdate)
	if err != nil {
		response := helper.BuildErrorResponse("Check your data!", err.Error(), helper.EmptyObj{})
		context.JSON(http.StatusForbidden, response)
	} else {
		res := helper.BuildResponse(true, "OK!", u)
		context.JSON(http.StatusOK, res)
	}
}

func (c *userController) All(context *gin.Context) {
	authHeader := context.GetHeader("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		panic(errToken.Error())
	}
	claims := token.Claims.(jwt.MapClaims)
	userID := fmt.Sprintf("%v", claims["user_id"])
	if c.userService.UserRole(userID) != "admin" {
		response := helper.BuildErrorResponse("You dont have permission", "You are not Administrator", helper.EmptyObj{})
		context.JSON(http.StatusForbidden, response)
	} else {
		var users []entity.User = c.userService.All()
		res := helper.BuildResponse(true, "OK", users)
		context.JSON(http.StatusOK, res)
	}
}

func (c *userController) Profile(context *gin.Context) {
	authHeader := context.GetHeader("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)

	if errToken != nil {
		panic(errToken.Error())
	}
	claims := token.Claims.(jwt.MapClaims)
	id := fmt.Sprintf("%v", claims["user_id"])
	user := c.userService.Profile(id)
	res := helper.BuildResponse(true, "OK!", user)
	context.JSON(http.StatusOK, res)
}
