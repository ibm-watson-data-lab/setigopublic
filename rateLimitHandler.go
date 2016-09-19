package main

import (
	"log"
	"net/http"
	"time"
)

func RateLimitHandler(inner http.Handler, rateLimited bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rateLimited {
			inner.ServeHTTP(w, r)
		} else {
			log.Printf("\n")
			log.Printf("Request for rate-limited resource.")
			token := r.URL.Query().Get("token")
			if token == "" {
				ReturnError(w, 500, "token_error", "Provide a token with your request '?token=abcdef123456789'")
				return
			}
			var rateLimitedResourceResponse, httpStatusCode, error, reason = VerifyRateLimitedResource(token, r.URL.String())
			if rateLimitedResourceResponse == nil {
				ReturnError(w, httpStatusCode, error, reason)
			} else {
				inner.ServeHTTP(w, r)
				// Record access in goroutine
				start := time.Now()
				go RecordRateLimitedRequest(rateLimitedResourceResponse.Token, rateLimitedResourceResponse.Request, rateLimitedResourceResponse.Date)
				log.Printf("Time to async POST to Cloudant: %s", time.Since(start))
			}
		}
	})
}
