package proxy

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func Proxy(prefix string) gin.HandlerFunc {

	return func(c *gin.Context) {
		//Setting up a proxy connection to octoprint
		remote, err := url.Parse("http://192.168.1.157:80")
		// remote, err := url.Parse("http://localhost:8000")
		if err != nil {
			panic(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(remote)

		//Modifying the request sent to the Proxy
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Path = func() string {
				var first string
				var second string
				var third string
				var fourth string
				var fith string

				//Proccessing each direcotry in path individually. This is the first
				if c.Param("first") == "" {
					return prefix
				}
				if c.Param("first")[0:1] != "/" {

					first = "/" + c.Param("first")
				} else {
					first = c.Param("first")
				}

				//This is the start of the second.
				if c.Param("second") == "" {
					return prefix + first
				}
				if c.Param("second")[0:1] != "/" {
					second = "/" + c.Param("second")
				} else {
					second = c.Param("second")
				}

				//This is the start of the Third
				if c.Param("third") == "" {
					return prefix + first + second
				}
				if c.Param("third")[0:1] != "/" {
					third = "/" + c.Param("third")
				} else {
					third = c.Param("third")
				}

				//This is the start of the fourth
				if c.Param("fourth") == "" {
					return prefix + first + second + third
				}
				if c.Param("fourth")[0:1] != "/" {
					fourth = "/" + c.Param("fourth")
				} else {
					fourth = c.Param("fourth")
				}

				//This is the start of the fith
				if c.Param("fith") == "" {
					return prefix + first + second + third + fourth
				}
				if c.Param("fith")[0:1] != "/" {
					fith = "/" + c.Param("fith")
				} else {
					fith = c.Param("fith")
				}
				return prefix + first + second + third + fourth + fith

			}()
			req.URL.RawQuery = c.Request.URL.RawQuery
			req.URL, err = url.Parse(remote.Scheme + "://" + remote.Host + req.URL.Path + "?" + c.Request.URL.RawQuery)
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("Trying to access %v on the proxy", req.URL)
			}

		}

		//Modify the response so that links/redirects work
		proxy.ModifyResponse = func(resp *http.Response) (err error) {

			//Correcting The response body so that href links work
			b, err := ioutil.ReadAll(resp.Body) //Read html
			if err != nil {
				log.Println(err)
			}
			err = resp.Body.Close()
			if err != nil {
				log.Println(err)
			}
			b = bytes.Replace(b, []byte("href=\"https://"), []byte("bref=\""), -1)
			b = bytes.Replace(b, []byte("href=\"/"), []byte("href=\"/proxy/"), -1)
			b = bytes.Replace(b, []byte("href=\""+remote.String()), []byte("href=\""+c.Request.URL.Scheme+"://"+c.Request.URL.Host+"proxy/"), -1) // replace html
			b = bytes.Replace(b, []byte("bref=\""), []byte("href=\"https://"), -1)
			body := ioutil.NopCloser(bytes.NewReader(b))
			resp.Body = body

			//Correcting The response location for redirects
			location, err := resp.Location()
			if err == nil && location.String() != "" {
				newLocation := location.String()
				newLocation = strings.Replace(newLocation, remote.String(), c.Request.URL.Scheme+c.Request.URL.Host+"/proxy", -1)
				resp.Header.Set("location", newLocation)
				log.Printf("Response is redirecting from %v and now to %v", location, newLocation)
			}
			resp.ContentLength = int64(len(b))
			resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
			return nil
		}
		//Serve content that was modified
		proxy.ServeHTTP(c.Writer, c.Request)
	}

}
