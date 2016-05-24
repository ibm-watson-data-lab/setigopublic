package main


import (
    "fmt"
    "net/http"
    "os"
		"log"
)

const (
	DEFAULT_PORT = "8080"
)


func main() {
  
  var port string
  if port = os.Getenv("PORT"); len(port) == 0 {
    port = DEFAULT_PORT
  }

  router := NewRouter()
  
  fmt.Println("listening on port "+port+" ...")
  
  log.Fatal(http.ListenAndServe(":"+port, router))
    
}

