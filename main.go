package main

import (
	"flag"
	"net/http"
	"os"
    "path/filepath"

	"github.com/caddyserver/certmagic"
	"github.com/gin-gonic/gin"
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
	password := flag.String("password", "", "Choose the app password obtained form no-reply email account")
	flag.Parse()

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

	err = certmagic.HTTPS([]string{"masongarten.sytes.net"}, router)
	if err != nil{
		panic(err)
	}

}
