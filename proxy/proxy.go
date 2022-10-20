package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	cat "github.com/masong19hippows/go-website/catError"
)

//declare a proxy type that holds the prefix for the url to access it
//as well as the url of the device being proxied to
type Proxy struct {
	AccessPrefix  string `json:"accessPrefix"`
	ProxyURL      string `json:"proxyURL"`
	AccessPostfix string `json:"accessPostfix"`
}

//array of all proxies
var Proxies []Proxy

// make sure Proxies is the latest
func init() {
	reloadProxies()

}

// refreshes Proxies with proxies.json
func reloadProxies() {
	Proxies = nil
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	jsonFile, err := os.Open(path.Join(exPath, "proxy", "web", "proxies.json"))
	if err != nil {
		log.Println(err)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &Proxies)
	Proxies = append(Proxies, Proxy{AccessPrefix: "/proxy/", ProxyURL: "http://localhost:6000", AccessPostfix: ""})
	jsonFile.Close()

}

//This is the middleware that handles the dynamic selection of proxies
func Handler(c *gin.Context) {

	log.Printf("Client requested %v", c.Request.URL)
	//Only pass if the error is 404
	if c.Writer.Status() == http.StatusNotFound {
		// Reloading list of proxies to make sure that the latest is used

		reloadProxies()
		//Getting the first directory in the url and matching it with prefixes in Proxies
		allSlash := regexp.MustCompile(`/(.*?)/`)

		prefix := allSlash.FindString(c.Request.URL.Path)
		if prefix == "" {
			prefix = c.Request.URL.Path + "/"
		}
		var final Proxy
		for _, proxy := range Proxies {
			if proxy.AccessPrefix == prefix {
				final = proxy
				break
			}
		}

		//Only pass if there is no proxy associated with the fist directory.
		//If this happens, the prefix for each proxy in the list is tested to see if it exists on the proxy
		//If it exists on the proxy, then traffic is redirected to Proxy
		//If it doesn't exist on the proxy, then a 404 is sent with a picture of a cat
		if (Proxy{}) == final {

			// Loop through proxies and find one that mjatches the prefix
			for i, proxy := range Proxies {

				requestURL := proxy.ProxyURL + func() string {
					if proxy.AccessPostfix == "" {
						return ""
					} else {
						return proxy.AccessPostfix[:len(proxy.AccessPostfix)-1]
					}
				}() + c.Request.URL.Path

				resp, err := http.Get(requestURL)
				if err != nil {
					log.Println(err)
					continue
				} else if resp.StatusCode == 404 {
					if i == len(Proxies)-1 {
						cat.SendError(cat.Response{Status: http.StatusNotFound, Error: []string{"File Not Found on Server"}}, c)
						break
					}
					continue
				} else {
					log.Printf("Proxy %v is redirecting traffic from %v to %v", proxy, c.Request.URL.Path, proxy.AccessPrefix[:len(proxy.AccessPrefix)-1]+c.Request.URL.Path)
					c.Redirect(http.StatusMovedPermanently, proxy.AccessPrefix[:len(proxy.AccessPrefix)-1]+c.Request.URL.Path)
				}

			}

		} else {

			//Look up the directory in the proxy
			lookProxy(final, c)
		}
	}

	// move on to other handlers
	c.Next()

}

// look up the url on the proxy. Send a 404 cat if not found
func lookProxy(lookup Proxy, c *gin.Context) {

	//Setting up a proxy connection to octoprint
	remote, err := url.Parse(lookup.ProxyURL)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)

	//Modifying the request sent to the Proxy
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host

		path := strings.Replace(c.Request.URL.Path, lookup.AccessPrefix, "", -1)
		if path == lookup.AccessPrefix[:len(lookup.AccessPrefix)-1] {
			path = ""
		}

		req.URL, err = url.Parse(remote.Scheme + "://" + remote.Host + func() string {
			if len(path) > 0 {
				if string(path[0]) == "/" {
					return ""
				} else {
					return "/"
				}
			} else {
				return ""
			}
		}() + lookup.AccessPostfix + path + func() string {
			if c.Request.URL.RawQuery == "" {
				return ""
			} else {
				return "?" + c.Request.URL.RawQuery
			}
		}())

		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Trying to access %v with the proxy %v", req.URL, lookup)
		}
	}

	//Modify the response so that links/redirects work
	proxy.ModifyResponse = func(resp *http.Response) (err error) {
		//Filter out the proxy reverse manager unless its from an internal ip address
		if lookup.AccessPrefix == "/proxy"{
			host, _, err := net.SplitHostPort(resp.Request.RemoteAddr)
			if err != nill{
				log.Println(err)
			} else {
				ip := net.ParseIP(host)
				if ip.IsPrivate() == false{
					cat.SendError(cat.Response{Status: http.StatusNotFound, Error: []string{"Not a Private IP Address"}}, c)
				}
			}
		}
		log.Printf("&v", resp.Request.RemoteAddr)
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
		b = bytes.Replace(b, []byte("href=\"/"), []byte("href=\""+lookup.AccessPrefix), -1)
		b = bytes.Replace(b, []byte("href=\""+remote.String()), []byte("href=\""+c.Request.URL.Scheme+"://"+c.Request.URL.Host+lookup.AccessPrefix), -1) // replace html
		b = bytes.Replace(b, []byte("bref=\""), []byte("href=\"https://"), -1)
		body := ioutil.NopCloser(bytes.NewReader(b))
		resp.Body = body

		//Correcting The response location for redirects
		location, err := resp.Location()
		if err == nil && location.String() != "" {
			newLocation := location.String()
			newLocation = strings.Replace(newLocation, remote.String(), c.Request.URL.Scheme+c.Request.URL.Host+lookup.AccessPrefix[:len(lookup.AccessPrefix)-1], -1)
			resp.Header.Set("location", newLocation)
			log.Printf("Response is redirecting from %v and now to %v", location, newLocation)
		}
		if resp.StatusCode == 404 {
			cat.SendError(cat.Response{Status: http.StatusNotFound, Error: []string{"File Not Found on Server"}}, c)

		}
		resp.ContentLength = int64(len(b))
		resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
		return nil
	}

	//Serve content that was modified
	proxy.ServeHTTP(c.Writer, c.Request)

}
