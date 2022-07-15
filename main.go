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
	//Setting up a proxy connection to octoprint
	remote, err := url.Parse("http://192.168.1.157:80")
	// remote, err := url.Parse("http://localhost:8000")
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)

	//Modifying the request sent to the Proxy
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Path = func() string {
			var first string
			var second string
			var third string
			var fourth string
			var fith string

			//Proccessing each direcotry in path individually. This is the first
			if c.Param("first") == "" {
				return "/"
			}
			if c.Param("first")[0:1] != "/" {

				first = "/" + c.Param("first")
			} else {
				first = c.Param("first")
			}
			if strings.Contains(first, ".") {
				return first
			}

			//This is the start of the second.
			if c.Param("second") == "" {
				return first + "/"
			}
			if c.Param("second")[0:1] != "/" {
				second = "/" + c.Param("second")
			} else {
				second = c.Param("second")
			}
			if strings.Contains(second, ".") {
				return first + second
			}

			//This is the start of the Third
			if c.Param("third") == "" {
				return first + second + "/"
			}
			if c.Param("third")[0:1] != "/" {
				third = "/" + c.Param("third")
			} else {
				third = c.Param("third")
			}
			if strings.Contains(third, ".") {
				return first + second + third
			}

			//This is the start of the fourth
			if c.Param("fourth") == "" {
				return first + second + third + "/"
			}
			if c.Param("fourth")[0:1] != "/" {
				fourth = "/" + c.Param("fourth")
			} else {
				fourth = c.Param("fourth")
			}
			if strings.Contains(fourth, ".") {
				return first + second + third + fourth
			}

			//This is the start of the fith
			if c.Param("fith") == "" {
				return first + second + third + fourth + "/"
			}
			if c.Param("fith")[0:1] != "/" {
				fith = "/" + c.Param("fith")
			} else {
				fith = c.Param("fith")
			}
			return first + second + third + fourth + fith

		}()
		req.URL.RawQuery = c.Request.URL.RawQuery
		req.URL, err = url.Parse(remote.Scheme + "://" + remote.Host + req.URL.Path + "?" + c.Request.URL.RawQuery)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Trying to access %v on the proxy", req.URL)
		}

	}

	//Modify the response so that links/redirects work
	proxy.ModifyResponse = func(resp *http.Response) (err error) {

		//Correcting The response body so that href links work
		b, err := ioutil.ReadAll(resp.Body) //Read html
		if err != nil {
			log.Println(err)
		}
		err = resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
		b = bytes.Replace(b, []byte("href=\"https://"), []byte("bref=\""), -1)
		b = bytes.Replace(b, []byte("href=\"/"), []byte("href=\"/octo/"), -1)
		b = bytes.Replace(b, []byte("href=\""+remote.String()), []byte("href=\""+c.Request.URL.Scheme+"://"+c.Request.URL.Host+"octo/"), -1) // replace html
		b = bytes.Replace(b, []byte("bref=\""), []byte("href=\"https://"), -1)
		fmt.Println(string(b)) // replace html
		body := ioutil.NopCloser(bytes.NewReader(b))
		resp.Body = body

		//Correcting The response location for redirects
		location, err := resp.Location()
		if err == nil && location.String() != "" {
			newLocation := location.String()
			newLocation = strings.Replace(newLocation, remote.String(), c.Request.URL.Scheme+c.Request.URL.Host+"/octo", -1)
			resp.Header.Set("location", newLocation)
			log.Printf("Response is redirecting from %v and now to %v", location, newLocation)
		}
		resp.ContentLength = int64(len(b))
		resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
		return nil
	}

	//Serve content that was modified
	proxy.ServeHTTP(c.Writer, c.Request)
}
func main() {

	port := flag.Int("port", 80, "Select the port that you wish the server to run on")
	password := flag.String("password", "", "Choose the app password obtained form no-reply email account")
	flag.Parse()
	log.Println("Using port", *port)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.NoMethod(SendError(Response{Status: http.StatusMethodNotAllowed, Error: []string{"File Not Found on Server"}}))
	router.NoRoute(SendError(Response{Status: http.StatusNotFound, Error: []string{"File Not Found on Server"}}))
	router.Any("/octo", proxy)
	router.Any("/octo/:first", proxy)
	router.Any("/octo/:first/:second", proxy)
	router.Any("/octo/:first/:second/:third", proxy)
	router.Any("/octo/:first/:second/:third/:fourth", proxy)
	router.Any("/octo/:first/:second/:third/:fourth/:fith", proxy)

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
