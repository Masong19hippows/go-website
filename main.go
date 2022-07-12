package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"

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

func sendEmail(password string) gin.HandlerFunc {

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
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> alert('Failed to send email. Error: "+e+"'); </script> </html>"))
		} else {
			log.Println("Mail sent successfully!")
			c.Status(http.StatusNoContent)
		}
	}

}
func proxy(c *gin.Context) {
	remote, err := url.Parse("http://192.168.1.157:80")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	fmt.Println(c.Request.Header)
	fmt.Println(remote.Host)
	fmt.Println(remote.Scheme)
	fmt.Println(c.Param("proxyPath"))

	//Define the director func
	//This is a good place to log, for example
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("proxyPath")
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func main() {

	port := flag.Int("port", 80, "Select the port that you wish the server to run on")
	password := flag.String("password", "", "Choose the app password obtained form no-reply email account")
	flag.Parse()
	log.Println("Using port", *port)

	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	router.NoMethod(SendError(Response{Status: http.StatusMethodNotAllowed, Error: []string{"File Not Found on Server"}}))
	router.NoRoute(SendError(Response{Status: http.StatusNotFound, Error: []string{"File Not Found on Server"}}))
	router.Any("/octo", proxy)

	router.StaticFile("/", "assets/index.html")
	router.POST("/send_email", sendEmail(*password))
	router.StaticFile("/favicon.ico", "assets/favicon.ico")
	router.StaticFile("/index.css", "assets/index.css")
	router.StaticFS("/images", http.Dir("./assets/images/"))

	err := router.Run(":" + strconv.Itoa(*port))
	if err != nil {
		log.Println(err)
	}
}
