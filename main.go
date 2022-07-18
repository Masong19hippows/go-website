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

	port := flag.Int("port", 80, "Select the port that you wish the server to run on")
	password := flag.String("password", "", "Choose the app password obtained form no-reply email account")
	flag.Parse()
	log.Println("Using port", *port)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(proxy.CreateAndReload())
	// router.NoMethod(cat.SendError(cat.Response{Status: http.StatusMethodNotAllowed, Error: []string{"No Method"}}))
	// router.NoRoute(cat.SendError(cat.Response{Status: http.StatusNotFound, Error: []string{"File Not Found on Server"}}))
	router.Any("/proxy", proxy.Proxy(""))
	router.Any("/proxy/:first", proxy.Proxy(""))
	router.Any("/proxy/:first/:second", proxy.Proxy(""))
	router.Any("/proxy/:first/:second/:third", proxy.Proxy(""))
	router.Any("/proxy/:first/:second/:third/:fourth", proxy.Proxy(""))
	router.Any("/proxy/:first/:second/:third/:fourth/:fith", proxy.Proxy(""))

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
