package main

import (
	// "errors"
	// "bytes"
	// "encoding/json"
	// "io"
	"io/ioutil"
	"log"
	"net/http"
	// "reflect"
	// "syscall/js"
)

// type Message struct {
// 		message string
// }

func main() {

	  requestUrl := "http://localhost:8080/leaderboard"
		req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
		if err != nil {
			log.Fatal(err)
		}

		// req.Header.Add("js.fetch:mode", "no-cors")
		// req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			// Handle errors: reject the Promise if we have an error
			log.Fatal(err)
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()

		log.Println("client payload", string(body))
}
