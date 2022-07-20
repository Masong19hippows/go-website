package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/masong19hippows/go-website/email"
	"github.com/masong19hippows/go-website/proxy"
)

func main() {

	//get port flag and password flag
	port := flag.Int("port", 80, "Select the port that you wish the server to run on")
	password := flag.String("password", "", "Choose the app password obtained form no-reply email account")
	flag.Parse()
	log.Println("Using port", *port, "For webserver")

	// non-verbose
	gin.SetMode(gin.ReleaseMode)

	//default routes + the proxy handler
	router := gin.Default()
	router.Use(proxy.Handler)
	router.StaticFile("/", "assets/index.html")
	router.POST("/send_email", email.SendEmail(*password))
	router.StaticFile("/favicon.ico", "assets/favicon.ico")
	router.StaticFile("/index.css", "assets/index.css")
	router.StaticFS("/images", http.Dir("./assets/images/"))

	err := router.Run(":" + strconv.Itoa(*port))
	if err != nil {
		log.Println(err)
	}
}
