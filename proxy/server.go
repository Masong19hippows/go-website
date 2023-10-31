package proxy

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	cat "github.com/masong19hippows/go-website/catError"
)

func init() {
	go server()
}

var temp []Proxy

func createProxy(webServer string, prefix string, postfix string, hostname bool) error {

	// Sanitizing the postfix by checking for whitespaces and "/"
	// at the end and beginning of string, if it exists
	postfix = strings.ReplaceAll(postfix, " ", "")
	if postfix != "" {
		if string(postfix[0]) != "/" {
			postfix = "/" + postfix
		}
		if string(postfix[len(postfix)-1]) != "/" {
			postfix = postfix + "/"
		}

	}

	// Sanitizing the prefix by checking for whitespaces and "/"
	// at the end and beginning of string
	prefix = strings.ReplaceAll(prefix, " ", "")

	if !hostname {
		if string(prefix[0]) != "/" {
			prefix = "/" + prefix
		}
		if string(prefix[len(prefix)-1]) != "/"{
			prefix = prefix + "/"
		}	

		// Sanitizing the url by checking for whitespaces
		// and checking if web server is reachable
		webServer = strings.ReplaceAll(webServer, " ", "")

		if string(webServer[len(webServer)-1]) == "/" {
			webServer = webServer[:len(webServer)-1]
		}
		fixedURL := webServer + postfix
		resp, err := http.Get(fixedURL)
		if err != nil {
			log.Println(err)
			return errors.New("cannot reach the URL: " + fixedURL)
		} else if resp.StatusCode == 404 {
			return errors.New("cannot reach url")
		} else if prefix == "" {
			return errors.New("no selection made")
		}

		// Check if its the same as any other Proxy
		for _, proxy := range Proxies {
			if prefix == proxy.AccessPrefix {
				return errors.New("prefix already exists: " + prefix)
			}
			if fixedURL == proxy.ProxyURL+proxy.AccessPostfix {
				return errors.New("URL already exists: " + fixedURL)
			}
		}
	}


	//Using the temporary
	temp = nil
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	jsonFile, err := os.Open(path.Join(exPath, "proxy", "web", "proxies.json"))
	if err != nil {
		jsonFile.Close()
		log.Println("Cannot open Proxies.json. Error is : ", err)
	} else {
		byteValue, _ := io.ReadAll(jsonFile)
		json.Unmarshal(byteValue, &temp)
		temp = append(temp, Proxy{AccessPrefix: prefix, ProxyURL: webServer, AccessPostfix: postfix, Hostname: hostname})
		jsonFile.Close()

	}

	//Append Changes to File
	result, err := json.Marshal(temp)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path.Join(exPath, "proxy", "web", "proxies.json"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	_, err = io.WriteString(f, string(result))
	if err != nil {
		return err
	}
	reloadProxies()
	log.Println("Proxies now contains: ", Proxies)

	return nil

}

func deleteProxy(index int) error {

	//Get new Slice to be ready to replace Proxies file with
	temp = nil
	temp = append(Proxies[:index], Proxies[index+1:]...)
	if len(temp) < 1 {
		log.Println("New Proxies is now", temp)
	} else {
		temp = temp[:len(temp)-1]
		log.Println("New Proxies is now", temp)
	}
	result, err := json.Marshal(temp)
	if err != nil {
		return err
	}

	//Append Changes to File
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)

	f, err := os.OpenFile(path.Join(exPath, "proxy", "web", "proxies.json"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	_, err = io.WriteString(f, string(result))
	if err != nil {
		return err
	}

	//Reload Proxies and Return with no error
	reloadProxies()
	log.Println("Proxies now contains: ", Proxies)

	return nil

}

func server() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.StaticFile("/", exPath+"/proxy/web/index.html")
	router.StaticFile("/proxy/index.html", exPath+"/proxy/web/index.html")
	router.StaticFile("/proxy", exPath+"/proxy/web/index.html")
	router.StaticFile("/proxy/proxies.json", exPath+"/proxy/web/proxies.json")
	router.StaticFile("/proxy/index.css", exPath+"/proxy/web/index.css")
	router.NoRoute(func(c *gin.Context) {
		cat.SendError(cat.Response{Status: http.StatusNotFound, Error: []string{"File Not Found on Server"}}, c)
	})

	router.POST("/proxy/create", func(c *gin.Context) {
		hostname := false
		if (c.PostForm("hostname") == "on"){
			hostname = true
		}
		err := createProxy(c.PostForm("url"), c.PostForm("prefix"), c.PostForm("postfix"), hostname)
		if err != nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> window.alert('Failed to Create Proxy. Error: "+err.Error()+"'); window.location.href='/proxy'; </script> </html>"))
		} else {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> window.alert('Succesfully Created Proxy'); window.location.href='/proxy'; </script> </html>"))
		}
	})

	router.GET("/proxy/delete/:index", func(c *gin.Context) {
		index, err := strconv.Atoi(c.Param("index"))
		if err != nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> window.alert('Failed to Delete Proxy. Error: "+err.Error()+"'); window.location.href='/proxy'; </script> </html>"))
		} else {
			err = deleteProxy(index)
			if err != nil {
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> window.alert('Failed to Delete Proxy. Error: "+err.Error()+"'); window.location.href='/proxy'; </script> </html>"))
			} else {
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> window.alert('Sucesfully Deleted Proxy'); window.location.href='/proxy'; </script> </html>"))
			}
		}
	})
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	err = router.Run("0.0.0.0:6000")
	if err != nil {
		panic(err)
	}
}
