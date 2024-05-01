package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"

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
	password := flag.String("password", "", "Choose the app password obtained form no-reply email account.")
	httpPort := flag.String("HTTPPort", "80", "Choose the HTTP Port.")
	proxyEnable := flag.Bool("proxy", true, "Enable to disable proxy.")
	flag.Parse() 

	// non-verbose
	gin.SetMode(gin.ReleaseMode)

	//default routes + the proxy handler
	router := gin.New()
	if *proxyEnable {
		proxy.Init()
		router.Use(proxy.Handler)
	}
	router.StaticFile("/", exPath+"/assets/index.html")
	router.POST("/send_email", email.SendEmail(*password))
	router.StaticFile("/favicon.ico", exPath+"/assets/favicon.ico")
	router.StaticFile("/index.css", exPath+"/assets/index.css")
	router.StaticFS("/images", http.Dir(exPath+"/assets/images/"))

	ch := make(chan error)
	go func(ch chan error) {
		err := router.Run("0.0.0.0:" + *httpPort)
		ch <- err
	}(ch)

	if *proxyEnable {
		go func(ch chan error) {

			var list []string
			var proxies []proxy.Proxy
			proxy.GetProxies(&proxies)
			list = append(list, "masongarten.com")
			for _, proxy := range proxies {
				if proxy.Hostname {
					list = append(list, proxy.AccessPrefix + ".masongarten.com")
				}
			}
			
			m := autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(list...),
				Cache:      autocert.DirCache(exPath + "/certs"),
			}
			err := autotls.RunWithManager(router, &m)
			ch <- err
		}(ch)
	
	}
	
	panic(<-ch)
}
