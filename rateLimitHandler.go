package main

import (
	"net/http"
)

func RateLimitHandler(inner http.Handler, rateLimited bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rateLimited {
			inner.ServeHTTP(w, r)
		} else {

			token := r.URL.Query().Get("access_token")
			if token == "" {
				ReturnError(w, 500, "access_token_error", "Provide an access_token with your request '?access_token=abcdef123456789'")
				return
			}
			var rateLimitedResourceResponse, httpStatusCode, error, reason = VerifyRateLimitedResource(token, r.URL.String())
			if rateLimitedResourceResponse == nil {
				ReturnError(w, httpStatusCode, error, reason)
			} else {
				inner.ServeHTTP(w, r)

				// Record access in goroutine
				go RecordRateLimitedRequest(rateLimitedResourceResponse.Token, rateLimitedResourceResponse.Request, rateLimitedResourceResponse.Date)

			}
		}
	})
}
