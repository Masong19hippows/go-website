package email

import (
	"log"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/gin-gonic/gin"
)

type EmailRequestBody struct {
	Name    string
	Email   string
	Message string
}

func SendEmail(password string) gin.HandlerFunc {

	return func(c *gin.Context) {
		email := c.PostForm("Email")
		name := c.PostForm("Name")
		message := c.PostForm("Message")
		username := "noreplymasongarten"
		smtpHost := "smtp.gmail.com"
		from := email

		to := []string{
			"garten323@gmail.com",
			email,
		}

		msg := []byte("From: " + email + "\r\n" +
			"Subject: Message from " + name + "\r\n\r\n" +
			"This email was generate by website. Please do not reply to this email. Message originally from " + email + "\n\n\n" +
			message + "\r\n")

		auth := smtp.PlainAuth("", username, password, smtpHost)
		err := smtp.SendMail(smtpHost+":587", auth, from, to, msg)

		if err != nil {

			var e string = err.Error()
			if strings.Contains(e, "Username and Password not accepted") {
				e = "Bad App Password/ App Username"
			}
			log.Println(err)
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> window.alert('Failed to Send Email. Error: "+e+"'); window.location.href='/'; </script> </html>"))
		} else {
			log.Println("Mail sent successfully!")
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> window.alert('Succesfully Sent Email'); window.location.href='/'; </script> </html>"))
		}
	}

}
