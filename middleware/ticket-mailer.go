package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nashirkra/Ticketing-Applications/entity"
	"github.com/nashirkra/Ticketing-Applications/service"
	"github.com/shopspring/decimal"
)

type MailMessage struct {
	Name          string
	TitleEvent    string
	UrlPay        string
	EventType     string
	EventDate     string
	StatusPayment string
	Price         string
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func LogResponseBody(c *gin.Context) {
	w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
	c.Writer = w
	c.Next()
	var trx map[string]interface{}
	err := json.Unmarshal(w.body.Bytes(), &trx)
	if err != nil {
		fmt.Printf("Error: %+v", err)
	}
	var data entity.Transaction
	jsonString, _ := json.Marshal(trx["data"].(map[string]interface{}))

	err = json.Unmarshal(jsonString, &data)
	if err != nil {
		fmt.Printf("Error2: %+v", err)
	}
	if data.ID > 0 {

		jwtServ := service.NewJWTService()
		generatedToken := jwtServ.GeneratePaymentToken(strconv.Itoa(data.ParticipantId), strconv.Itoa(data.ID), "Processing")
		dataToMail := MailMessage{
			Name:          data.Participant.Fullname,
			TitleEvent:    data.Event.TitleEvent,
			UrlPay:        "http://localhost:8080/api/payment/" + generatedToken,
			EventType:     data.Event.TypeEvent,
			EventDate:     data.Event.EventEndDate.Format(time.RFC3339),
			StatusPayment: data.StatusPayment,
			Price:         decimal.NewFromFloat(data.Amount).String(),
		}
		// go service.SendMailTrx(data.Participant.Email, dataToMail)
		go service.SendMailTrx("1379c0b91c-3e80d1@inbox.mailtrap.io", dataToMail)
	} else {
		fmt.Println("Data null!")
	}
}
