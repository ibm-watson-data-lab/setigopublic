package main

import "net/http"

func AuthHandler(inner http.Handler, authRequired bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !authRequired {
			inner.ServeHTTP(w, r)
		} else {
			// if user is not logged in (no user in session) then redirect to login page
			session, err := sessionStore.Get(r, "user_session")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			user := session.Values["user"]
			if user == nil {
				http.Redirect(w, r, "/login?r="+r.RequestURI, http.StatusTemporaryRedirect)
			} else {
				inner.ServeHTTP(w, r)
			}
		}
	})
}
