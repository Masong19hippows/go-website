package main

import (
	"flag"
	"log"
	"fmt"
	"net/http"
	"strconv"
	"os"
    "path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/masong19hippows/go-website/certs"
	"github.com/masong19hippows/go-website/email"
	"github.com/masong19hippows/go-website/proxy"
)

func main() {
	ex, err := os.Executable()
    if err != nil {
        panic(err)
    }
    exPath := filepath.Dir(ex)

	//get port flag and password flag
	portHTTP := flag.Int("portHTTP", 80, "Select the port that you wish the http server to run on")
	portHTTPS := flag.Int("portHTTPS", 443, "Select the port that you wish the https server to run on")
	password := flag.String("password", "", "Choose the app password obtained form no-reply email account")
	flag.Parse()
	log.Println("Using port", *portHTTP, "For http webserver and port", *portHTTPS, "for https server")

	// non-verbose
	gin.SetMode(gin.ReleaseMode)

	//default routes + the proxy handler
	router := gin.New()
	router.Use(proxy.Handler)
	router.StaticFile("/", exPath + "/assets/index.html")
	router.POST("/send_email", email.SendEmail(*password))
	router.StaticFile("/favicon.ico", exPath + "/assets/favicon.ico")
	router.StaticFile("/index.css", exPath + "/assets/index.css")
	router.StaticFS("/images", http.Dir(exPath + "/assets/images/"))

	ch := make(chan error)
	go func (ch chan error) {
		err := router.Run("0.0.0.0:" + strconv.Itoa(*portHTTP))
		ch <- err
		return
	}(ch)
	go func (ch chan error) {
		err := certs.Gen(exPath)
		if err != nil{
			fmt.Println(err)
			ch <- err
			return
		}
		err = http.ListenAndServeTLS(":" + strconv.Itoa(*portHTTPS), exPath + "/certs/cert.pem", exPath + "/certs/key.pem", nil)
		ch <- err
		return
	}(ch)

	<-ch
}
