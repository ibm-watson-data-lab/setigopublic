package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"bytes"
	"log"
	"os"
	"time"
)

var (
	httpClient *http.Client
)

const (
	MaxIdleConnections int = 20
	RequestTimeout     int = 10
)

type CloudantViewRow struct {
	Key   string `json:"key"`
	Id    string `json:"id"`
	Value uint64 `json:"value"`
}

type CloudantResponse struct {
	TotalRows uint64            `json:"total_rows"`
	Offset    uint64            `json:"offset"`
	Rows      []CloudantViewRow `json:"rows"`
}

type CloudantUserDoc struct {
	UserID string `json:"user_id"`
	Limit  int64  `json:"limit"`
}

type CloudantSuccessfulPost struct {
	Ok  bool   `json:"ok"`
	Id  string `json:"id"`
	Rev string `json:"rev"`
}

type GetTokenResponse struct {
	AccessToken string `json:"acces_token"` //this will be the same as the doc _id for the user's doc in Cloudant
}

type ViewRow struct {
	Key   []string `json:"key"`
	Value uint64   `json:"value"`
}

type ViewResult struct {
	Rows []ViewRow `json:"rows"`
}

type RateLimitedResourceResponse struct {
	Token   string
	Request string
	Date    int64
}

func getHttpClient() *http.Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: MaxIdleConnections,
				DisableKeepAlives:   false,
			},
			Timeout: time.Duration(RequestTimeout) * time.Second,
		}
	}
	return httpClient
}

func GetToken(user *User) (getTokenResponse *GetTokenResponse, httpStatusCode int, error string, reason string) {

	cloudant_base_url := os.Getenv("CLOUDANT_SETI_USERS_BASE_URL")
	user_id_view := "_design/docs/_view/user_id"

	//The user has been logged in via Bluemix at this point
	//1. Check our Cloudant database to see if this user has already
	//   been issued a token (which will be just the doc._id)

	user_url := cloudant_base_url + user_id_view + "?key=\"" + user.UserID + "\""
	client := getHttpClient()

	response, err := client.Get(user_url)
	defer response.Body.Close()
	if err != nil {
		log.Printf("cloudant_error: %s", err.Error())
		return nil, 500, "cloudant_error", "unable to GET user_id view"
	}

	//very simple doc to hold user id information
	cloudant_response := CloudantResponse{}
	err = json.NewDecoder(response.Body).Decode(&cloudant_response)

	if err != nil {
		log.Printf("json_decode_error. %s", err.Error())
		return nil, 500, "json_decode_error", "Could not parse user_id view Return"
	}

	if len(cloudant_response.Rows) > 1 {
		return nil, 500, "cloudant_error", "too many rows!"
	}

	if len(cloudant_response.Rows) == 0 {
		//this user doesn't exist in the DB.
		log.Printf("Didn't find User. Creating new user in database")

		new_user := CloudantUserDoc{UserID: user.UserID, Limit: 0}

		//post to cloudant
		//return access_token in JSON
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(new_user)

		response, err := client.Post(cloudant_base_url, "application/json", b)
		ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			log.Printf("cloudant_error. %s", err.Error())
			return nil, 500, "cloudant_error", "Error POSTing new user doc"
		}

		//decode the response to get the "id", then create a ReturnData object

		var success_post CloudantSuccessfulPost
		err = json.NewDecoder(response.Body).Decode(&success_post)

		if err != nil {
			log.Printf("json_decode_error. %s", err.Error())
			return nil, 500, "json_decode_error", "Failed to decode POST response"
		}

		return &GetTokenResponse{AccessToken: success_post.Id}, 400, "", ""

	} else {
		//found the user
		log.Print("Found user. Returning access token.")
		user_row := cloudant_response.Rows[0]

		return &GetTokenResponse{AccessToken: user_row.Id}, 400, "", ""
	}
}

func VerifyRateLimitedResource(token string, url string) (rateLimitedResourceResponse *RateLimitedResourceResponse, httpStatusCode int, error string, reason string) {
	//Check to see if document exists. The token should be the doc ID
	cloudant_base_url := os.Getenv("CLOUDANT_SETI_USERS_BASE_URL")
	limit, _ := strconv.ParseUint(os.Getenv("SETI_USERS_MONTHLY_LIMIT"), 10, 64)

	start := time.Now()
	req_url := cloudant_base_url + token
	client := getHttpClient()

	response, err := client.Get(req_url)
	defer response.Body.Close()
	if err != nil {
		log.Printf("cloudant_error: %s", err.Error())
		return nil, 500, "cloudant_error", "Error GETing token document"
	}

	log.Printf("Time to GET document: %s", time.Since(start))
	start = time.Now()

	if response.StatusCode == 200 {
		var doc CloudantUserDoc
		err = json.NewDecoder(response.Body).Decode(&doc)
		if err != nil {
			log.Printf("json_decode_error. %s", err.Error())
			return nil, 500, "json_decode_error", "Could not parse token_date_size view results"
		}
		if doc.Limit < 0 {
			limit = ^uint64(0)
		} else if doc.Limit > 0 {
			limit = uint64(doc.Limit)
		}
	} else {
		return nil, 500, "cloudant_error", "Token Not Found."
	}

	log.Printf("Limit for token %s = %d", token, limit)

	//Now check to see if the number of documents requested so
	//far for this token has exceeded our monthly limit
	now := time.Now()
	date_today := now.Unix()
	date_last_month := now.AddDate(0, -1, 0).Unix()

	//size view url
	size_view_string := fmt.Sprintf("_design/docs/_view/token_date_size?startkey=[\"%s\",%d]&endkey=[\"%s\",%d]", token, date_last_month, token, date_today)

	log.Printf("View URL: %s", size_view_string)
	req_url = cloudant_base_url + size_view_string
	response, err = client.Get(req_url)
	defer response.Body.Close()
	if err != nil {
		log.Printf("cloudant_error: %s", err.Error())
		return nil, 500, "cloudant_error", "Error GETing token access count"
	}

	log.Printf("Time to GET view results: %s", time.Since(start))
	start = time.Now()

	var view_results ViewResult
	err = json.NewDecoder(response.Body).Decode(&view_results)

	if err != nil {
		log.Printf("json_decode_error. %s", err.Error())
		return nil, 500, "json_decode_error", "Could not parse token_date_size view results"
	}

	//The size of Rows *should* be only 1 or 0
	if len(view_results.Rows) > 1 {
		log.Printf("cloudant_error. Unexpected size of token_data_size view results")
		return nil, 500, "cloudant_error", "Unexpected size of token_data_size view results"
	}

	if len(view_results.Rows) == 1 {
		if view_results.Rows[0].Value >= limit {
			log.Printf("data_limit for %s", token)
			return nil, 403, "data_limit_error", "You've reached your limit for the month."
		}
	}
	log.Printf("Time to decode and perform logic: %s", time.Since(start))
	
	return &RateLimitedResourceResponse{Token: token, Request: url, Date: date_today}, 400, "", ""
}

// Record rate-limited request for user in Cloudant
func RecordRateLimitedRequest(token string, request string, date int64) {
	type CloudantDataDoc struct {
		AccessToken string `json:"access_token"`
		Request     string `json:"request"`
		Date        int64  `json:"date"`
	}

	new_doc := CloudantDataDoc{
		AccessToken: token,
		Request:     request,
		Date:        date,
	}

	//post to cloudant
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(new_doc)

	cloudant_base_url := os.Getenv("CLOUDANT_SETI_USERS_BASE_URL")
	client := getHttpClient()
	response, err := client.Post(cloudant_base_url, "application/json", b)
	ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Printf("cloudant_error. Error POSTing new user doc. %s", err.Error())
	}
}
