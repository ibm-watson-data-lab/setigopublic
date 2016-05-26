package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
  cfenv "github.com/cloudfoundry-community/go-cfenv"
  "strconv"
)

const (
	DEFAULT_PORT = "8080"
)

func main() {

  var port string
  fmt.Println("Searching for PORT")

  appEnv, err := cfenv.Current()
  if err != nil {
    //we are not in a CF environment. Attempt to get PORT from local envars
  	if port = os.Getenv("PORT"); len(port) == 0 {
		  port = DEFAULT_PORT
	 }
  } else {
    port = strconv.Itoa(appEnv.Port)
  }

	router := NewRouter()

	fmt.Println("listening on port " + port + " ...")

	log.Fatal(http.ListenAndServe(":"+port, router))

}
