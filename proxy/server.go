package proxy

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	cat "github.com/masong19hippows/go-website/catError"
)

func init() {
	go server()
}

var temp []Proxy

func createProxy(webServer string, prefix string, postfix string) error {

	//Checking for correct input and trying to correct if possible
	if postfix != "" {
		if string(postfix[0]) != "/" {
			postfix = "/" + postfix
		}
		if string(postfix[len(postfix)-1]) != "/" {
			postfix = postfix + "/"
		}
	}
	if string(prefix[0]) != "/" {
		prefix = "/" + prefix
	}
	if string(prefix[len(prefix)-1]) != "/" {
		prefix = prefix + "/"
	}

	//Check if web server is reachable
	resp, err := http.Get(webServer + postfix)
	if err != nil {
		log.Println(err)
		return errors.New("Cannot reach the URL: " + webServer + postfix)
	} else if resp.StatusCode == 404 {
		return errors.New("cannot reach url")
	} else if prefix == "" {
		return errors.New("no selection made")
	}

	//Check if its the same as any other Proxy
	for _, proxy := range Proxies {
		if prefix == proxy.AccessPrefix {
			return errors.New("prefix already exists: " + prefix)
		}
		if webServer+postfix == proxy.ProxyURL+proxy.AccessPostfix {
			return errors.New("URL already exists: " + webServer + postfix)
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
		log.Println("Successfully Opened proxies.json")
		byteValue, _ := ioutil.ReadAll(jsonFile)
		json.Unmarshal(byteValue, &temp)
		temp = append(temp, Proxy{AccessPrefix: prefix, ProxyURL: webServer, AccessPostfix: postfix})
		jsonFile.Close()

	}

	//Append Changes to File
	result, err := json.Marshal(temp)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path.Join(exPath, "proxy", "web", "proxies.json"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Println("test", err)
		return err
	}

	_, err = io.WriteString(f, string(result))
	if err != nil {
		return err
	}
	reloadProxies()
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
		log.Println("test", err)
		return err
	}

	_, err = io.WriteString(f, string(result))
	if err != nil {
		return err
	}

	//Reload Proxies and Return with no error
	reloadProxies()
	return nil

}

func server() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.StaticFile("/", "proxy/web/index.html")
	router.StaticFile("/proxy", "proxy/web/index.html")
	router.StaticFile("/proxies.json", "proxy/web/proxies.json")
	router.StaticFile("/proxy/proxies.json", "proxy/web/proxies.json")
	router.StaticFile("/index.css", "proxy/web/index.css")
	router.StaticFile("/proxy/index.css", "proxy/web/index.css")
	router.NoRoute(func(c *gin.Context) {
		cat.SendError(cat.Response{Status: http.StatusNotFound, Error: []string{"File Not Found on Server"}}, c)
	})

	router.POST("/create", func(c *gin.Context) {
		err := createProxy(c.PostForm("url"), c.PostForm("prefix"), c.PostForm("postfix"))
		if err != nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> window.alert('Failed to Create Proxy. Error: "+err.Error()+"'); window.location.href='/proxy'; </script> </html>"))
		} else {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><script> window.alert('Succesfully Created Proxy'); window.location.href='/proxy'; </script> </html>"))
		}
	})

	router.GET("/delete/:test", func(c *gin.Context) {
		index, err := strconv.Atoi(c.Param("test"))
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

	err := router.Run(":5000")
	if err != nil {
		panic(err)
	}
}
