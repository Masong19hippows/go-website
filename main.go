package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
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
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> window.alert('Failed to Send Email. Error: "+e+"'); window.location.href='/'; </script> </html>"))
		} else {
			log.Println("Mail sent successfully!")
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> window.alert('Succesfully Sent Email'); window.location.href='/'; </script> </html>"))
		}
	}

}
func proxy(c *gin.Context) {
	remote, err := url.Parse("http://192.168.1.157:80")
	// remote, err := url.Parse("http://localhost:8000")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("octo") + "/" + c.Param("test")
	}

	proxy.ModifyResponse = func(resp *http.Response) (err error) {
		b, err := ioutil.ReadAll(resp.Body) //Read html
		if err != nil {
			log.Println(err)
		}
		err = resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
		b = bytes.Replace(b, []byte("href=\""), []byte("href=\"/octo"), -1) // replace html
		body := ioutil.NopCloser(bytes.NewReader(b))
		resp.Body = body

		test, err := resp.Location()
		if err != nil {
			log.Println(err)
		} else {
			test1 := test.String()
			test1 = strings.Replace(test1, "http://192.168.1.157:80/", "/octo/", -01)
			resp.Header.Set("location", test1)
			fmt.Println(test1)
		}

		test1 := test.String()
		test1 = strings.Replace(test1, "http://192.168.1.157:80/", "/octo/", -01)
		resp.Header.Set("Location", test1)
		fmt.Println(test1)

		resp.ContentLength = int64(len(b))
		resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
		return nil
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
	router.Any("/octo/:octo", proxy)
	router.Any("/octo/:octo/:test", proxy)

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
