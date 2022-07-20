package cat

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status int
	Error  []string
}

// use http cat api to get cute cat errors
func downloadError(error int) []byte {
	//Get the response bytes from the url
	cat, err := http.Get("https://http.cat/" + strconv.Itoa(error))
	if err != nil {
		log.Fatal(err)
	}

	defer cat.Body.Close()
	result, err := io.ReadAll(cat.Body)
	if err != nil {
		log.Fatal(err)
	}

	return result
}

// send the error back using gin.context
func SendError(response Response, c *gin.Context) {

	c.Data(response.Status, "image/png", downloadError(response.Status))

}
