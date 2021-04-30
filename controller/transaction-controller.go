package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/nashirkra/Ticketing-Applications/helper"
	"github.com/nashirkra/Ticketing-Applications/service"
	"github.com/nashirkra/Ticketing-Applications/valueObjects"
)

type TransactionController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
}

type transactionController struct {
	transationService service.TransactionService
	userService       service.UserService
	jwtService        service.JWTService
}

func NewTransactionController(trxServ service.TransactionService, userServ service.UserService, jwtServ service.JWTService) TransactionController {
	return &transactionController{
		transationService: trxServ,
		userService:       userServ,
		jwtService:        jwtServ,
	}
}

func (c *transactionController) Create(ctx *gin.Context) {
	var createdVO valueObjects.Transaction

	err := ctx.ShouldBindJSON(&createdVO)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	} else {
		authHeader := ctx.GetHeader("Authorization")
		token, errToken := c.jwtService.ValidateToken(authHeader)
		if errToken != nil {
			panic(errToken.Error())
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := fmt.Sprintf("%v", claims["user_id"])

		createdVO.CreatorId, _ = strconv.Atoi(userID)
		result, err := c.transationService.Create(createdVO)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to Create Transaction", err.Error(), helper.EmptyObj{})
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		response := helper.BuildResponse(true, "OK", result)
		ctx.JSON(http.StatusCreated, response)
	}
}

func (c *transactionController) Update(ctx *gin.Context) {
	var updateVO valueObjects.Transaction

	err := ctx.ShouldBindJSON(&updateVO)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	} else {
		authHeader := ctx.GetHeader("Authorization")
		token, errToken := c.jwtService.ValidateToken(authHeader)
		if errToken != nil {
			panic(errToken.Error())
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := fmt.Sprintf("%v", claims["user_id"])

		updateVO.CreatorId, _ = strconv.Atoi(userID)
		result, err := c.transationService.Update(updateVO)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to Create Transaction", err.Error(), helper.EmptyObj{})
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		response := helper.BuildResponse(true, "OK", result)
		ctx.JSON(http.StatusCreated, response)
	}
}
