package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status int
	Error  []string
}

type EmailRequestBody struct {
	Name    string
	Email   string
	Message string
}

func downloadError(error int) []byte {
	//Get the response bytes from the url
	cat, err := http.Get("https://http.cat/" + strconv.Itoa(error))
	if err != nil {
		log.Fatal(err)
	}

	defer cat.Body.Close()
	result, err := io.ReadAll(cat.Body)
	if err != nil {
		log.Fatal(err)
	}

	return result
}

func SendError(response Response) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(response.Status, "image/png", downloadError(response.Status))
	}
}

func sendEmail(c *gin.Context) {

	email := c.PostForm("Email")
	name := c.PostForm("Name")
	message := c.PostForm("Message")
	username := "garten323"
	appPassword := "nope"
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

	auth := smtp.PlainAuth("", username, appPassword, smtpHost)
	err := smtp.SendMail(smtpHost+":587", auth, from, to, msg)

	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Mail sent successfully!")

		c.Status(http.StatusNoContent)
	}

}

func main() {

	router := gin.Default()

	router.NoMethod(SendError(Response{Status: http.StatusMethodNotAllowed, Error: []string{"File Not Found on Server"}}))
	router.NoRoute(SendError(Response{Status: http.StatusNotFound, Error: []string{"File Not Found on Server"}}))

	router.StaticFile("/", "assets/index.html")
	router.POST("/send_email", sendEmail)
	router.StaticFile("/favicon.ico", "assets/favicon.ico")
	router.StaticFile("/index.css", "assets/index.css")
	router.StaticFS("/images", http.Dir("./assets/images/"))

	err := router.Run(":80")
	if err != nil {
		fmt.Println(err)
	}
}
