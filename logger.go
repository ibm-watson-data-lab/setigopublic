package main

import (
	"log"
	"net/http"
	"time"
	"net/url"
	"os"
	"encoding/json"
	"bytes"
)


func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)

		//post to cloudant
		go func(r *http.Request){
			clouurl := os.Getenv("CLOUDANT_URL")
		
			if clouurl == "" {
				return
			}

			type LogData struct {
		    URL url.URL `json:"url"`
		    Host string `json:"host"`
		    RemoteAddr string `json:"remote_address"`
		    Type string `json:"type"`
		    DocVersion int `json:"doc_version"`
		    Date string `json:"date"`
		    Unix int64 `json:"date_epoch"`
		    AppVersion string `json:"app_version"`
		    Route string `json:app_route`
		  }

		  nn := time.Now()

		  returnData := LogData{ 
		  	URL:*r.URL, 
		  	Host:r.Host, 
		  	RemoteAddr:r.RemoteAddr,
		  	Type:"go-callback",
		  	DocVersion:1,
		  	Date:nn.String(),
		  	Unix:nn.Unix(),
				AppVersion:APP_VERSION,
				Route:name}  //APP_VERSION found in app.go

		  buf, _ := json.Marshal(returnData)

			req, _ := http.NewRequest("POST", clouurl, bytes.NewBuffer(buf))
	    req.Header.Set("Content-Type", "application/json")

	    client := &http.Client{}
	    resp, err := client.Do(req)
	    if err != nil {
	        panic(err)
	    }
	    defer resp.Body.Close()

		}(r)

	})
}
