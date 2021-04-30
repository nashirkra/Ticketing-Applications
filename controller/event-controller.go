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
	"github.com/shopspring/decimal"
)

type TransactionPost struct {
	ParticipantId string `json:"participant_id"`
	CreatorId     string `json:"creator_id"`
	EventId       string `json:"event_id"`
	Amount        string `json:"amount"`
	StatusPayment string `json:"status_payment"`
}

type EventController interface {
	Create(ctx *gin.Context)
	GetEvent(ctx *gin.Context)
	AllEvent(ctx *gin.Context)
}

type eventController struct {
	eventService service.EventService
	jwtService   service.JWTService
}

func NewEventController(evServ service.EventService, jwtServ service.JWTService) EventController {
	return &eventController{
		eventService: evServ,
		jwtService:   jwtServ,
	}
}

func (c *eventController) Create(ctx *gin.Context) {
	var createdVO valueObjects.Event

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
		userRole := c.eventService.UserRole(userID)
		if (userRole == "admin") || (userRole == "creator") {
			createdVO.CreatorId, _ = strconv.Atoi(userID)
			result, err := c.eventService.Create(createdVO)
			if err != nil {
				response := helper.BuildErrorResponse("Failed to Create Event", err.Error(), helper.EmptyObj{})
				ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
				return
			}
			response := helper.BuildResponse(true, "OK", result)
			ctx.JSON(http.StatusCreated, response)
		} else {
			response := helper.BuildErrorResponse("You dont have permission", "You are not Administrator", helper.EmptyObj{})
			ctx.JSON(http.StatusForbidden, response)
		}
	}
}

func (c *eventController) GetEvent(ctx *gin.Context) {
	eventID, err := strconv.ParseUint(ctx.Param("id"), 0, 0)
	if err != nil {
		res := helper.BuildErrorResponse("No param id was found", err.Error(), helper.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	event := c.eventService.GetEvent(strconv.FormatUint(eventID, 10))

	authHeader := ctx.GetHeader("Authorization")
	if len(authHeader) > 0 {
		token, errToken := c.jwtService.ValidateToken(authHeader)

		if errToken != nil {
			panic(errToken.Error())
		}
		claims := token.Claims.(jwt.MapClaims)
		participantId, _ := strconv.Atoi(fmt.Sprintf("%v", claims["user_id"]))
		trxObj := TransactionPost{
			ParticipantId: strconv.Itoa(participantId),
			CreatorId:     strconv.Itoa(event.CreatorId),
			EventId:       strconv.Itoa(event.ID),
			Amount:        decimal.NewFromFloat(event.Price).String(),
			StatusPayment: "Pending",
		}
		res := helper.BuildResponseEmbed(true, "OK!", event, trxObj)
		ctx.JSON(http.StatusOK, res)
	} else {
		res := helper.BuildResponse(true, "OK!", event)
		ctx.JSON(http.StatusOK, res)
	}
}
func (c *eventController) AllEvent(ctx *gin.Context) {
	events := c.eventService.AllEvent()

	res := helper.BuildResponse(true, "OK!", events)
	ctx.JSON(http.StatusOK, res)
}
