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
	"gorm.io/gorm"
)

type PaymentController interface {
	InfoPayment(ctx *gin.Context)
	Pay(ctx *gin.Context)
	Cancel(ctx *gin.Context)
}

type paymentController struct {
	transationService service.TransactionService
	jwtService        service.JWTService
}

func NewPaymentController(trxServ service.TransactionService, jwtServ service.JWTService) PaymentController {
	return &paymentController{
		transationService: trxServ,
		jwtService:        jwtServ,
	}
}

func (c *paymentController) InfoPayment(ctx *gin.Context) {
	ticketToken := ctx.Param("token")
	if len(ticketToken) == 0 {
		res := helper.BuildErrorResponse("No param token was found", "Error!", helper.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	token, err := c.jwtService.ValidateToken(ticketToken)
	if err != nil {
		res := helper.BuildErrorResponse("Token Failed", err.Error(), helper.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	res := helper.BuildResponse(true, "OK! Create our info payment", token)
	ctx.JSON(http.StatusOK, res)
}
func (c *paymentController) Pay(ctx *gin.Context) {
	var updateVO valueObjects.Transaction
	ticketToken := ctx.Param("token")
	if len(ticketToken) == 0 {
		res := helper.BuildErrorResponse("No param token was found", "Error!", helper.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	// xtract token
	token, err := c.jwtService.ValidateToken(ticketToken)
	if err != nil {
		res := helper.BuildErrorResponse("Token Failed", err.Error(), helper.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	// userID := fmt.Sprintf("%v", claims["user_id"])
	ticketID := fmt.Sprintf("%v", claims["ticket_id"])
	statusDo := fmt.Sprintf("%v", claims["status"])
	if statusDo == "Processing" {
		getTicket, err := c.transationService.GetTicket(ticketID)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to Get Transaction", err.Error(), helper.EmptyObj{})
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}

		updateVO.ParticipantId = getTicket.ParticipantId
		updateVO.EventId = getTicket.EventId
		updateVO.ID, _ = strconv.Atoi(ticketID)
		updateVO.StatusPayment = "Completed"
		result, err := c.transationService.Update(updateVO)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to Update Transaction", err.Error(), helper.EmptyObj{})
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		response := helper.BuildResponse(true, "OK! assumed the payment was Completed", result)
		ctx.JSON(http.StatusCreated, response)
	}
}
func (c *paymentController) Cancel(ctx *gin.Context) {
	var updateVO valueObjects.Transaction
	ticketToken := ctx.Param("token")
	if len(ticketToken) == 0 {
		res := helper.BuildErrorResponse("No param token was found", "Error!", helper.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	// xtract token
	token, err := c.jwtService.ValidateToken(ticketToken)
	if err != nil {
		res := helper.BuildErrorResponse("Token Failed", err.Error(), helper.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	// userID := fmt.Sprintf("%v", claims["user_id"])
	ticketID := fmt.Sprintf("%v", claims["ticket_id"])
	statusDo := fmt.Sprintf("%v", claims["status"])
	// Do verify current status here

	if statusDo == "Processing" {
		getTicket, err := c.transationService.GetTicket(ticketID)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to Get Transaction", err.Error(), helper.EmptyObj{})
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		updateVO.ParticipantId = getTicket.ParticipantId
		updateVO.EventId = getTicket.EventId
		updateVO.ID, _ = strconv.Atoi(ticketID)
		updateVO.StatusPayment = "Cancelled"
		updateVO.DeletedAt = gorm.DeletedAt{Time: helper.TimeIn("Jakarta"), Valid: true}
		result, err := c.transationService.Update(updateVO)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to Update Transaction", err.Error(), helper.EmptyObj{})
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		response := helper.BuildResponse(true, "OK! assumed the payment procedure was canceled", result)
		ctx.JSON(http.StatusCreated, response)
	}
}
