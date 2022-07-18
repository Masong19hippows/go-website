package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	cat "github.com/masong19hippows/go-website/catError"
	"github.com/masong19hippows/go-website/email"
	"github.com/masong19hippows/go-website/proxy"
)

func createAndReload() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get file path and check if exists
		// If not, create
		// No need to redirect as FS fill pick it up now

		// continue with the flow
		c.Next()

		// 404 will never happen
		status := c.Writer.Status()
		if status == 404 {
			newPath := c.Request.URL.Scheme + c.Request.URL.Host + "/proxy" + c.Request.URL.Path
			fmt.Println(newPath)
			c.Redirect(http.StatusMovedPermanently, newPath)
		}
	}
}

func main() {

	port := flag.Int("port", 80, "Select the port that you wish the server to run on")
	password := flag.String("password", "", "Choose the app password obtained form no-reply email account")
	flag.Parse()
	log.Println("Using port", *port)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(createAndReload())
	router.NoMethod(cat.SendError(cat.Response{Status: http.StatusMethodNotAllowed, Error: []string{"No Method"}}))
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
