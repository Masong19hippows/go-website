package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/masong19hippows/go-website/email"
	"github.com/masong19hippows/go-website/proxy"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	//get port flag and password flag
	portHTTP := flag.Int("portHTTP", 80, "Select the port that you wish the http server to run on")
	password := flag.String("password", "", "Choose the app password obtained form no-reply email account")
	flag.Parse()
	log.Println("Using port", *portHTTP, "For http webserver and port 443 for https server")

	// non-verbose
	gin.SetMode(gin.ReleaseMode)

	//default routes + the proxy handler
	router := gin.New()
	router.Use(proxy.Handler)
	router.StaticFile("/", exPath+"/assets/index.html")
	router.POST("/send_email", email.SendEmail(*password))
	router.StaticFile("/favicon.ico", exPath+"/assets/favicon.ico")
	router.StaticFile("/index.css", exPath+"/assets/index.css")
	router.StaticFS("/images", http.Dir(exPath+"/assets/images/"))

	ch := make(chan error)
	go func(ch chan error) {
		err := router.Run("0.0.0.0:" + strconv.Itoa(*portHTTP))
		ch <- err
	}(ch)
	go func(ch chan error) {
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("masongarten.com"),
			Cache:      autocert.DirCache(exPath + "/certs"),
		}
		err := autotls.RunWithManager(router, &m)
		ch <- err
	}(ch)

	panic(<-ch)
}
