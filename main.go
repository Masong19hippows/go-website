package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	cat "github.com/masong19hippows/go-website/catError"
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
	router.NoMethod(cat.SendError(cat.Response{Status: http.StatusMethodNotAllowed, Error: []string{"File Not Found on Server"}}))
	router.NoRoute(cat.SendError(cat.Response{Status: http.StatusNotFound, Error: []string{"No Method"}}))
	router.Any("/octo", proxy.Proxy(""))
	router.Any("/octo/:first", proxy.Proxy(""))
	router.Any("/octo/:first/:second", proxy.Proxy(""))
	router.Any("/octo/:first/:second/:third", proxy.Proxy(""))
	router.Any("/octo/:first/:second/:third/:fourth", proxy.Proxy(""))
	router.Any("/octo/:first/:second/:third/:fourth/:fith", proxy.Proxy(""))

	router.Any("/static", proxy.Proxy("/static"))
	router.Any("/static/:first", proxy.Proxy("/static"))
	router.Any("/static/:first/:second", proxy.Proxy("/static"))
	router.Any("/static/:first/:second/:third", proxy.Proxy("/static"))
	router.Any("/static/:first/:second/:third/:fourth", proxy.Proxy("/static"))
	router.Any("/static/:first/:second/:third/:fourth/:fith", proxy.Proxy("/static"))

	router.Any("/sockjs", proxy.Proxy("/sockjs"))
	router.Any("/sockjs/:first", proxy.Proxy("/sockjs"))
	router.Any("/sockjs/:first/:second", proxy.Proxy("/sockjs"))

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
