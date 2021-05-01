package service

import (
	"bytes"
	"context"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/go-redis/redis/v8"
)

func SendMailTrx(mailto string, data interface{}) {
	// Sender data.
	from := "d9864df36db270"
	password := "8adb942df9b33a"

	// Receiver email address.
	to := []string{
		mailto,
	}

	// smtp server configuration.
	smtpHost := "smtp.mailtrap.io"
	smtpPort := "2525"

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	t, _ := template.ParseFiles("template.html")

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("Subject: Pemesanan Ticket \n%s\n\n", mimeHeaders)))

	t.Execute(&body, data)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
	if err != nil {
		fmt.Println(err)
		fmt.Println(data)
		return
	}
	fmt.Println("Email Sent!")
}

func DailyMail() {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// sample list of mail
	mail, err := client.RPopLPush(ctx, "dailyMail", "dailyMail").Result()
	if err != nil {
		fmt.Printf("Error Sending Mail: %+v\n", err)
	}
	fmt.Printf("Sending Mail: %+v\n", mail)
}
