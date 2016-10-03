package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"bytes"

	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/ncw/swift"
	"golang.org/x/oauth2"
)


var oauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
	Scopes:       []string{"openid"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://login.ng.bluemix.net/UAALoginServerWAR/oauth/authorize",
		TokenURL: "https://login.ng.bluemix.net/UAALoginServerWAR/oauth/token",
	},
}
var sessionStore = sessions.NewCookieStore([]byte("$ECRET$ETIcode"))

// Logout
// This clears the user from the session
// It does not log the user out from Bluemix
func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, "user_session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["user"] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// Login
// This redirects the user to the Bluemix Login page
func Login(w http.ResponseWriter, r *http.Request) {
	// redirect to Bluemix login page
	redirect := r.URL.Query().Get("r")
	url := oauthConfig.AuthCodeURL(redirect, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Auth Handler
// After logging into Bluemix a user is redirected to this Endpoint
// with the authorization code assigned by Bluemix
// After a successful login this function queries Bluemix for the users's info
func BluemixAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConfig.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	// request userinfo from Bluemix
	oauthClient := oauthConfig.Client(oauth2.NoContext, token)
	response, err := oauthClient.Get("https://login.ng.bluemix.net/UAALoginServerWAR/userinfo")
	body, _ := ioutil.ReadAll(response.Body)
	// save user in session
	session, err := sessionStore.Get(r, "user_session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user := User{}
	json.Unmarshal(body, &user)
	session.Values["user"] = user
	session.Save(r, w)
	log.Printf("USER = %s", user.Email)
	// redirect to Index
	redirect := r.FormValue("state")
	if redirect == "" {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

func Token(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, "user_session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	val := session.Values["user"]
	var user = &User{}
	user, ok := val.(*User)
	if !ok {
		ReturnError(w, 500, "session_error", "you are not logged in")
		return
	}
	var getTokenResponse, httpStatusCode, error, reason = GetToken(user)
	if getTokenResponse == nil {
		ReturnError(w, httpStatusCode, error, reason)
	} else {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(getTokenResponse); err != nil {
			panic(err)
		}
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	// load user from session
	session, err := sessionStore.Get(r, "user_session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	val := session.Values["user"]
	var user = &User{}
	user, ok := val.(*User)
	// print out html
	fmt.Fprintln(w, "<html><head><title>SETI on IBM-Spark</title></head><body>")
	fmt.Fprintln(w, "<p><h3>Welcome to SETI Public on Spark.</h3></p>")
	fmt.Fprintln(w, "<p>Brought to you by IBM, the SETI Institute in Mountain View, CA and NASA.</p>")
	if !ok {
		fmt.Fprintln(w, "<p><a href='/login'>Login with Bluemix</a></p>")
	} else {
		fmt.Fprintln(w, "<p>You are logged in as "+user.UserName+"</p>")
	}
	fmt.Fprintln(w, "</body></html>")
}

func AcaByCoordinates(w http.ResponseWriter, r *http.Request) {

	var coordinates CelestialCoordinates
	var err error = nil

	vars := mux.Vars(r)

	coordinates.RA, err = strconv.ParseFloat(vars["ra"], 64)
	if err != nil {
		ReturnError(w, 400, "parse_error", "Unable to parse RA value.")
		return
	}

	coordinates.Dec, err = strconv.ParseFloat(vars["dec"], 64)
	if err != nil {
		ReturnError(w, 400, "parse_error", "Unable to parse DEC value.")
		return
	}

	log.Printf("coordinates: %v", coordinates)

	//use this to allow for a query to skip a number of initial rows
	//we limit the output of this query to a maximum of 200 rows per query
	skiprows, _ := strconv.ParseInt(r.URL.Query().Get("skip"), 10, 64)

	//use this to allow for a query to limit the number of returned rows.
	//however, the maximum allowed is 200 rows per query
	var limit int64 = 200
	if r.URL.Query().Get("limit") != "" {
		limit, _ = strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
		if limit > 200 {
			limit = 200
		}
	}


	var totalNumRows int64

	row := dbConnection.QueryRow(`SELECT count(*) FROM (SELECT UNIQUEID FROM SETIUSERS.SIGNALDB WHERE SIGCLASS='Cand'
    AND RA2000HR = ? AND DEC2000DEG = ?) as SDB 
    INNER JOIN  SETIUSERS.SDB_PATH_TO_ACA AS ACA 
    ON SDB.UNIQUEID = ACA.UNIQUEID`, coordinates.RA, coordinates.Dec)

	err = row.Scan(&totalNumRows)
	if err != nil {
		ReturnError(w, 500, "query_count_error", err.Error())
		return
	}

	signalDBJoinACAPaths := []SignalDBJoinACAPath{}

	err = dbConnection.Select(&signalDBJoinACAPaths, `SELECT SDB.*, ACA.CONTAINER AS CONTAINER, ACA.OBJECTNAME AS OBJECTNAME
    FROM (SELECT * FROM SETIUSERS.SIGNALDB WHERE SIGCLASS='Cand' AND RA2000HR = ? AND DEC2000DEG = ?) as SDB 
    INNER JOIN  SETIUSERS.SDB_PATH_TO_ACA AS ACA 
    ON SDB.UNIQUEID = ACA.UNIQUEID 
    ORDER BY SDB.UNIQUEID 
    LIMIT ? OFFSET ?`, coordinates.RA, coordinates.Dec, limit, skiprows)

	//todo: GROUP BY OBJECTNAME -- so we only send unique object names?
	//OR should I group the results here instead?, allowing for multiple sets of signalDB rows to
	//be returned? Can probably do that within the SQL query

	if err != nil {
		ReturnError(w, 500, "query_rows_error", err.Error())
		return
	}

	type ReturnData struct {
		TotalNumRows int64                 `json:"total_num_rows"`
		Skip         int64                 `json:"skipped_num_rows"`
		Size         int                   `json:"returned_num_rows"`
		Data         []SignalDBJoinACAPath `json:"rows"`
	}

	returnData := ReturnData{TotalNumRows: totalNumRows, Skip: skiprows, Size: len(signalDBJoinACAPaths), Data: signalDBJoinACAPaths}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

func SpaceCraft(w http.ResponseWriter, r *http.Request) {

	var err error = nil

	//use this to allow for a query to skip a number of initial rows
	//we limit the output of this query to a maximum of 200 rows per query
	skiprows, _ := strconv.ParseInt(r.URL.Query().Get("skip"), 10, 64)

	//use this to allow for a query to limit the number of returned rows.
	//however, the maximum allowed is 200 rows per query
	var limit int64 = 200
	if r.URL.Query().Get("limit") != "" {
		limit, _ = strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
		if limit > 200 {
			limit = 200
		}
	}


	var totalNumRows int64

	row := dbConnection.QueryRow(`SELECT count(*) FROM (SELECT UNIQUEID FROM SETIUSERS.SIGNALDB WHERE CATALOG='spacecraft') as SDB 
    INNER JOIN  SETIUSERS.SDB_PATH_TO_ACA AS ACA 
    ON SDB.UNIQUEID = ACA.UNIQUEID`)

	err = row.Scan(&totalNumRows)
	if err != nil {
		ReturnError(w, 500, "query_count_error", err.Error())
		return
	}

	signalDBJoinACAPaths := []SignalDBJoinACAPath{}

	err = dbConnection.Select(&signalDBJoinACAPaths, `SELECT SDB.*, ACA.CONTAINER AS CONTAINER, ACA.OBJECTNAME AS OBJECTNAME
    FROM (SELECT * FROM SETIUSERS.SIGNALDB WHERE CATALOG='spacecraft') as SDB 
    INNER JOIN  SETIUSERS.SDB_PATH_TO_ACA AS ACA 
    ON SDB.UNIQUEID = ACA.UNIQUEID 
    ORDER BY SDB.UNIQUEID 
    LIMIT ? OFFSET ?`, limit, skiprows)

	if err != nil {
		ReturnError(w, 500, "query_rows_error", err.Error())
		return
	}

	type ReturnData struct {
		TotalNumRows int64                 `json:"total_num_rows"`
		Skip         int64                 `json:"skipped_num_rows"`
		Size         int                   `json:"returned_num_rows"`
		Data         []SignalDBJoinACAPath `json:"rows"`
	}

	returnData := ReturnData{TotalNumRows: totalNumRows, Skip: skiprows, Size: len(signalDBJoinACAPaths), Data: signalDBJoinACAPaths}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

func KnownCandCoordinates(w http.ResponseWriter, r *http.Request) {
	//query parameters
	//skip
	//limit
	//ramin
	//ramax
	//decmin
	//decmax

	//use this to allow for a query to skip a number of initial rows
	//we limit the output of this query to a maximum of 200 rows per query
	skiprows, _ := strconv.ParseInt(r.URL.Query().Get("skip"), 10, 64)

	//use this to allow for a query to limit the number of returned rows.
	//however, the maximum allowed is 200 rows per query
	var limit int64 = 200
	if r.URL.Query().Get("limit") != "" {
		limit, _ = strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
		if limit > 200 {
			limit = 200
		}
	}

	ramin := 0.0
	if r.URL.Query().Get("ramin") != "" {
		ramin, _ = strconv.ParseFloat(r.URL.Query().Get("ramin"), 64)
	}

	ramax := 24.0
	if r.URL.Query().Get("ramax") != "" {
		ramax, _ = strconv.ParseFloat(r.URL.Query().Get("ramax"), 64)
	}

	decmin := -90.0
	if r.URL.Query().Get("decmin") != "" {
		decmin, _ = strconv.ParseFloat(r.URL.Query().Get("decmin"), 64)
	}

	decmax := 90.0
	if r.URL.Query().Get("decmax") != "" {
		decmax, _ = strconv.ParseFloat(r.URL.Query().Get("decmax"), 64)
	}

	var totalNumRows int64

	row := dbConnection.QueryRow(`SELECT count(*) FROM SETIUSERS.ACA_CANDIDATE_COORDINATES 
    WHERE RA2000HR >= ? AND RA2000HR < ? AND DEC2000DEG >= ? AND DEC2000DEG < ?`, ramin, ramax, decmin, decmax)

	err := row.Scan(&totalNumRows)
	if err != nil {
		ReturnError(w, 500, "query_count_error", err.Error())
		return
	}

	knownACACoordinates := []KnownACACoordinate{}
	err = dbConnection.Select(&knownACACoordinates, `SELECT * FROM SETIUSERS.ACA_CANDIDATE_COORDINATES 
    WHERE RA2000HR >= ? AND RA2000HR < ? AND DEC2000DEG >= ? AND DEC2000DEG < ? 
    ORDER BY RA2000HR, DEC2000DEG LIMIT ? OFFSET ?`, ramin, ramax, decmin, decmax, limit, skiprows)

	if err != nil {
		ReturnError(w, 500, "query_rows_error", err.Error())
		return
	}

	type ReturnData struct {
		TotalNumRows int64                `json:"total_num_rows"`
		Skip         int64                `json:"skipped_num_rows"`
		Size         int                  `json:"returned_num_rows"`
		Data         []KnownACACoordinate `json:"rows"`
	}

	returnData := ReturnData{TotalNumRows: totalNumRows, Skip: skiprows, Size: len(knownACACoordinates), Data: knownACACoordinates}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

func getSetiPublicConnectionWithLocalEnvars() swift.Connection {

	c := swift.Connection{
		UserName: os.Getenv("SWIFT_API_USER"),      //username
		ApiKey:   os.Getenv("SWIFT_API_KEY"),       //password
		AuthUrl:  os.Getenv("SWIFT_AUTH_URL"),      //reponsibility of envar to contain the full URL, including v1.0, v2, or v3
		Domain:   os.Getenv("SWIFT_API_DOMAIN"),    //domainName (optional, for v3 only)
		DomainId: os.Getenv("SWIFT_API_DOMAIN_ID"), //domainId (optional, for v3 only)
		Tenant:   os.Getenv("SWIFT_TENANT"),        //project in vcap_services on bluemix (optional, for v3 only)
		TenantId: os.Getenv("SWIFT_TENANT_ID"),     //projectId in vcap_services on bluemix (optional, for v3 only)
	}
	return c
}

func getSetiPublicConnection() swift.Connection {
	var c swift.Connection

	appEnv, err := cfenv.Current()
	if err != nil {
		//we are not in a CF environment. Attempt to get credentials from local envars
		//we use the same envar names that are used in the swift library tests

		c = getSetiPublicConnectionWithLocalEnvars()

	} else {
		services, err := appEnv.Services.WithLabel("Object-Storage")

		if err != nil {
			//even though we're in a CF environgment, we didn't find an object store
			//bound to the app. So, we instantiate a swift.Connection object
			//with envars that should be set.
			c = getSetiPublicConnectionWithLocalEnvars()

		} else {
			objstore := services[0] //assume it's the only one (bad)
			c = swift.Connection{
				UserName: objstore.Credentials["userId"].(string),
				ApiKey:   objstore.Credentials["password"].(string),
				AuthUrl:  objstore.Credentials["auth_url"].(string) + "/v3", //have to add this manually here. no other way.
				Domain:   objstore.Credentials["domainName"].(string),
				DomainId: objstore.Credentials["domainId"].(string),
				Tenant:   objstore.Credentials["project"].(string),
				TenantId: objstore.Credentials["projectId"].(string),
			}
		}
	}

	return c
}

func GetACARawDataTempURL(w http.ResponseWriter, r *http.Request) {

	swift_secret_key := os.Getenv("SWIFT_SECRET_KEY")

	if swift_secret_key == "" {
		ReturnError(w, 500, "temp_url_error", "secret key not found")
		return
	}

	vars := mux.Vars(r)
	container := vars["container"]
	objectname := vars["date"] + "/" + vars["act"] + "/" + vars["acafile"]

	c := getSetiPublicConnection()

	err := c.Authenticate()
	if err != nil {
		ReturnError(w, 500, "temp_url_error", err.Error())
		return
	}

	//default to 1 hour... But check for envar
	//that can control expiration time
	//For example, coudl be set to "60s", or "24h".
	//See https://golang.org/pkg/time/#ParseDuration
	expiration := time.Now().Add(time.Second * 3600)

	if os.Getenv("EXPIRATION_TIME") != "" {
		extim, err := time.ParseDuration(os.Getenv("EXPIRATION_TIME"))
		if err == nil {
			expiration = time.Now().Add(extim)
		} else {
			fmt.Println("Failed to parse expiration from EXPIRATION_TIME: " + err.Error())
		}
	}

	temp_url := c.ObjectTempUrl(container, objectname, swift_secret_key, "GET", expiration)

	if temp_url == "" {
		ReturnError(w, 500, "temp_url_error", "returned empty URL")
		return
	}

	type ReturnData struct {
		Url    string `json:"temp_url"`
		Notice string `json:"license_notification"`
	}

	license := "This data is licensed by the SETI Institute under the Creative Commons BY 4.0 license.  https://github.com/ibm-cds-labs/seti_at_ibm/blob/master/setigopublic.md#data-license"
	returnData := ReturnData{Url: temp_url, Notice: license}

	retjson, err := json.Marshal(returnData);
	if  err != nil {
		panic(err)
	}

	//fix the encoding bug where '&' is decoded by the json.Marshal
  retjson = bytes.Replace(retjson, []byte("\\u0026"), []byte("&"), -1)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(retjson)

}

