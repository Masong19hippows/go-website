package proxy

import (
	"encoding/json"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"bytes"

	"github.com/gin-gonic/gin"
	cat "github.com/masong19hippows/go-website/catError"
)

// declare a proxy type that holds the prefix for the url to access it
// as well as the url of the device being proxied to
type Proxy struct {
	AccessPrefix  string `json:"accessPrefix"`
	ProxyURL      string `json:"proxyURL"`
	AccessPostfix string `json:"accessPostfix"`
	Hostname	  bool   `json:"hostname"`
	ForcePaths	  bool   `json:"forcepaths"`
	ReadHTML	  bool   `json:"readhtml"`
}


// array of all proxies
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
	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &Proxies)
	Proxies = append(Proxies, Proxy{AccessPrefix: "/proxy/", ProxyURL: "http://localhost:6000", AccessPostfix: "", Hostname: false, ForcePaths: true, ReadHTML: false})
	jsonFile.Close()

}

// This is the middleware that handles the dynamic selection of proxies
func Handler(c *gin.Context) {
	
	log.Printf("Client requested %v", c.Request.URL)

	if (c.Request.Host != "masongarten.com"){
		host_parts := strings.Split(c.Request.Host, ".")
		subdomain := host_parts[0]

		reloadProxies()
		
		var final Proxy
		for _, proxy := range Proxies {
			if !proxy.Hostname {
				continue
			}
			if proxy.AccessPrefix == subdomain {
				final = proxy
				break
			}
		}

		if (Proxy{}) == final {
			cat.SendError(cat.Response{Status: http.StatusNotFound, Error: []string{"File Not Found on Server"}}, c)
			return
		} else {
			//Look up the directory in the proxy
			lookProxy(final, c)
			return
		}
		

	}
	
	//Redirecting http to https
	if c.Request.TLS == nil {
		c.Redirect(http.StatusMovedPermanently, "https://"+c.Request.Host+c.Request.URL.Path+func() string {
			if c.Request.URL.RawQuery == "" {
				return ""
			} else {
				return "?" + c.Request.URL.RawQuery
			}
		}())
		return
	}

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
			if proxy.Hostname == true {
				continue
			}
			if proxy.AccessPrefix == prefix {
				final = proxy
				break
			}
		}

		//Only pass if there is no proxy associated with the directory.
		//If the proxy doesn't exist, then a 404 is sent with a picture of a cat
		if (Proxy{}) == final {
			cat.SendError(cat.Response{Status: http.StatusNotFound, Error: []string{"File Not Found on Server"}}, c)
			return
		} else {
			//Look up the directory in the proxy
			lookProxy(final, c)
			return
		}
	}
}

type copyWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (cw copyWriter) Write(b []byte) (int, error) {
	return cw.buf.Write(b)
}

// look up the url on the proxy. Send a 404 cat if not found
func lookProxy(lookup Proxy, c *gin.Context) {
	c.Abort()
	
	//Setting up a proxy connection
	remote, err := url.Parse(lookup.ProxyURL)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Transport = &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	cw := &copyWriter{buf: &bytes.Buffer{}, ResponseWriter: c.Writer}
	c.Writer = cw
	

	//Modifying the request sent to the Proxy
	proxy.Director = func(req *http.Request) {

		//Setting the connection up so it looks like its not form the Reverse Proxy Server
		req.Header = c.Request.Header
		
		ip, _, err := net.SplitHostPort(req.RemoteAddr)
    		if err != nil {
			log.Println(err)
    		}
		
		req.Header.Set("X-Forwarded-For", ip)
		req.Header.Set("X-Frame-Options", "SAMEORIGIN")
		req.Header.Set("X-Forwarded-Host", c.Request.Host)
		req.Header.Set("X-Scheme", remote.Scheme)
		req.Header.Set("X-Forwarded-Proto", remote.Scheme)
		req.Host = remote.Host
		ips, err := net.LookupIP(c.Request.Host)
		if err != nil {
			log.Fatal(err)
		}
		
		req.Header.Set("X-Real-IP", ips[0].String())
		
		

		if !lookup.Hostname {
			req.Header.Set("X-Script-Name", lookup.AccessPrefix[:len(lookup.AccessPrefix)-1])
		}

		path := c.Request.URL.Path


		if !lookup.Hostname {
			if lookup.ForcePaths {
				if path == lookup.AccessPrefix[:len(lookup.AccessPrefix)-1] {
					path += remote.Path
				}
			} else {
				path = strings.Replace(c.Request.URL.Path, lookup.AccessPrefix, "", -1)
				if path == lookup.AccessPrefix[:len(lookup.AccessPrefix)-1] {
					path = ""
				}
			}
		}

		

		req.URL, err = url.Parse(remote.Scheme + "://" + remote.Host + func() string {
			if len(path) > 0 {
				if string(path[0]) == "/" {
					return ""
				} else {
					return "/"
				}
			} else {
				return "/"
			}
		}() + func() string {
			if len(lookup.AccessPostfix) <= 1 {
				return lookup.AccessPostfix
			} else {
				return lookup.AccessPostfix[1:]
			}
		}() + path + func() string {
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
		// Returning 404 if getting a 404
		if resp.StatusCode == 404 {
			log.Println("got 404 with " + resp.Request.URL.String())
			cat.SendError(cat.Response{Status: http.StatusNotFound, Error: []string{"File Not Found on Server"}}, c)
			return nil

		}
		//Filter out the proxy reverse manager unless its from an internal ip address
		if lookup.AccessPrefix == "/proxy/" {
			host, _, err := net.SplitHostPort(resp.Request.RemoteAddr)
			if err != nil {
				log.Println(err)
			} else {
				ip := net.ParseIP(host)
				if !ip.IsPrivate() && ip.String() != "127.0.0.1" {
					log.Printf("Denied Acces to Proxy from %v", ip)
					cat.SendError(cat.Response{Status: http.StatusNotFound, Error: []string{"Not a Private IP Address"}}, c)
					return nil
				}
			}
		}

		//Correcting The response location for redirects

		location, err := resp.Location()
		if err == nil && location.String() != "" {
			newLocation := location.String()
			if !lookup.Hostname {
				newLocation =  strings.Replace(newLocation, "http://" + remote.Host, c.Request.URL.Scheme+c.Request.URL.Host, -1)
				newLocation =  strings.Replace(newLocation, "https://" + remote.Host, c.Request.URL.Scheme+c.Request.URL.Host, -1)
			} else {
				newLocation = strings.Replace(newLocation, remote.String(), c.Request.URL.Scheme+c.Request.URL.Host, -1)
				newLocation = "https://" + lookup.AccessPrefix + ".masongarten.com" + newLocation
			}
			newLocation = func() string {
				if lookup.AccessPostfix == "" {
					return newLocation
				}
				
				idx := strings.Index(newLocation, lookup.AccessPostfix)
				if newLocation[idx-1] == '/' {
					return strings.Replace(newLocation, lookup.AccessPostfix, "", -1)
				} else {
					return strings.Replace(newLocation, lookup.AccessPostfix, "/", -1)
				}
			}()
			resp.Header.Set("location", newLocation)
			log.Printf("Response from proxy is redirecting from %v and now to %v", location, newLocation)
			return nil
		}	

		if lookup.ReadHTML {
			b, err := io.ReadAll(resp.Body) //Read html
			defer resp.Body.Close()
			if err != nil {
				log.Fatalln(err)
			}

			b = bytes.Replace(b, []byte("href=\"https://"), []byte("bref=\""), -1)
			b = bytes.Replace(b, []byte("href=\"/"), []byte("href=\""+lookup.AccessPrefix), -1)
			b = bytes.Replace(b, []byte("href=\""+remote.String()), []byte("href=\""+c.Request.URL.Scheme+"://"+c.Request.URL.Host+lookup.AccessPrefix), -1) // replace html
			b = bytes.Replace(b, []byte("bref=\""), []byte("href=\"https://"), -1)

			b = bytes.Replace(b, []byte("src=\"https://"), []byte("bsrc=\""), -1)
			b = bytes.Replace(b, []byte("src=\"/"), []byte("src=\""+lookup.AccessPrefix), -1)
			b = bytes.Replace(b, []byte("src=\""+remote.String()), []byte("src=\""+c.Request.URL.Scheme+"://"+c.Request.URL.Host+lookup.AccessPrefix), -1) // replace html
			b = bytes.Replace(b, []byte("bsrc=\""), []byte("src=\"https://"), -1)
			body := io.NopCloser(bytes.NewReader(b))
			resp.Body = body
		}
		resp.Header.Set("X-Frame-Options", "SAMEORIGIN")


		return nil
	}
	//Serve content that was modified
	proxy.ServeHTTP(c.ResponseWriter, c.Request)

}
