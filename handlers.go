package main

import (
	_ "bitbucket.org/phiggins/db2cli"
	"encoding/json"
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"net/http"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os"
	"strconv"
	"strings"
  "fmt"
  "errors"
)

var dashDB, dashdbuser, dashdbpass = fetchCredentials()


func fetchCredentials() (cfenv.Service, string, string) {

  // Returns the dashdB Service, plus DASHDBUSER, DASHDBPASS

  //Fetch dashDB service details
  var dashDB cfenv.Service

  appEnv, err := cfenv.Current()
  if err != nil {
    //we are not in a CF environment. Attempt to get dashDB credentials from local envars
    vcap := os.Getenv("VCAP_SERVICES")

    var vcapj cfenv.Services

    if vcap == "" {
        panic(errors.New("No VCAP_SERVICES found."))
    }

    if err := json.Unmarshal([]byte(vcap), &vcapj); err != nil {
      panic(err)
    }

    services, err := vcapj.WithLabel("dashDB")

    if err != nil {
      panic(err)
    }

    dashDB = services[0]

  } else {
    services, err := appEnv.Services.WithLabel("dashDB")

    if err != nil {
      panic(err)
    }

    dashDB = services[0]
  }

  //I should probably use the setiusers values for username/password instead of the admin values
  //This might be safer.
  dashdbuser := os.Getenv("DASHDBUSER")
  dashdbpass := os.Getenv("DASHDBPASS")
  if dashdbuser == "" {
    panic(errors.New("No DASHDBUSER found."))
  }

  if dashdbpass == "" {
    panic(errors.New("No DASHDBPASS found."))
  }

  return dashDB, dashdbuser, dashdbpass
}

func Index(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintln(w, "<html><head><title>SETI on IBM-Spark</title></head><body>")
  fmt.Fprintln(w, "<p><h3>Welcome to SETI Public on Spark.</h3></p>")
  fmt.Fprintln(w, "<p>Brought to you by IBM, the SETI Institute in Mountain View, CA and NASA.</p>")
  fmt.Fprintln(w, "</body></html>")
}

func AcaByCoordinates(w http.ResponseWriter, r *http.Request) {

	var coordinates CelestialCoordinates
	var err error = nil

	coordinates.RA, err = strconv.ParseFloat(r.URL.Query().Get("RA"), 64)
	if err != nil {
		ReturnError(w, 400, "missing_data", "No Ra value.")
		return
	}

	coordinates.Dec, err = strconv.ParseFloat(r.URL.Query().Get("Dec"), 64)
	if err != nil {
		ReturnError(w, 400, "missing_data", "No Dec value.")
		return
	}

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

	

	connStr := []string{"DATABASE=", dashDB.Credentials["db"].(string), ";", "HOSTNAME=", dashDB.Credentials["hostname"].(string), ";",
		"PORT=", strconv.FormatFloat(dashDB.Credentials["port"].(float64), 'f', 0, 64), ";", "PROTOCOL=TCPIP", ";", "UID=", dashdbuser, ";", "PWD=", dashdbpass}
	conn := strings.Join(connStr, "")

	db, err := sqlx.Connect("db2-cli", conn)
	if err != nil {
		ReturnError(w, 500, "db2_error", "Unable to open connection.")
		return
	}
  db.MapperFunc(strings.ToUpper)
	defer db.Close()

	
  var totalNumRows int64

  row := db.QueryRow(`SELECT count(*) FROM (SELECT UNIQUEID FROM SETIUSERS.SIGNALDB WHERE SIGCLASS='Cand' AND RA2000HR < ? 
    AND RA2000HR > ? AND DEC2000DEG < ? AND DEC2000DEG > ?) as SDB 
    INNER JOIN  SETIUSERS.SDB_PATH_TO_ACA AS ACA 
    ON SDB.UNIQUEID = ACA.UNIQUEID`, coordinates.RA + 0.01, coordinates.RA - 0.01, coordinates.Dec + 0.01, coordinates.Dec - 0.01 )

  err = row.Scan(&totalNumRows)
  if err != nil {
    ReturnError(w, 500, "query_count_error", err.Error())
    return
  }

  signalDBJoinACAPaths := []SignalDBJoinACAPath{}

  err = db.Select(&signalDBJoinACAPaths, `SELECT SDB.*, ACA.CONTAINER AS CONTAINER, ACA.OBJECTNAME AS OBJECTNAME
    FROM (SELECT * FROM SETIUSERS.SIGNALDB WHERE SIGCLASS='Cand' AND RA2000HR < ? 
    AND RA2000HR > ? AND DEC2000DEG < ? AND DEC2000DEG > ?) as SDB 
    INNER JOIN  SETIUSERS.SDB_PATH_TO_ACA AS ACA 
    ON SDB.UNIQUEID = ACA.UNIQUEID 
    ORDER BY SDB.UNIQUEID 
    LIMIT ? OFFSET ?`, coordinates.RA + 0.01, coordinates.RA - 0.01, coordinates.Dec + 0.01, coordinates.Dec - 0.01, limit, skiprows )


  if err != nil {
    ReturnError(w, 500, "query_rows_error", err.Error())
    return
  }

  type ReturnData struct {
    TotalNumRows int64 `json:"total_num_rows"`
    Skip int64 `json:"skipped_num_rows"`
    Size int `json:"returned_num_rows"`
    Data []SignalDBJoinACAPath `json:"rows"`
  }

  returnData := ReturnData{TotalNumRows: totalNumRows, Skip:skiprows, Size:len(signalDBJoinACAPaths), Data:signalDBJoinACAPaths}


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
    if ramin < 0 {
      ramin = 0.0
    }  
  }
  
  ramax := 24.0
  if r.URL.Query().Get("ramax") != "" {
    ramax, _ = strconv.ParseFloat(r.URL.Query().Get("ramax"), 64)
    if ramax > 24 {
      ramax = 24.0
    }
  }
  decmin := -90.0
  if r.URL.Query().Get("decmin") != "" {
    decmin, _ = strconv.ParseFloat(r.URL.Query().Get("decmin"), 64)
    if decmin < -90 {
      decmin = -90.0
    }
  }
  decmax := 90.0
  if r.URL.Query().Get("decmax") != "" {
    decmax, _ := strconv.ParseFloat(r.URL.Query().Get("decmax"), 64)
    if decmax > 90 {
      decmax = 90.0
    }  
  }
  

  connStr := []string{"DATABASE=", dashDB.Credentials["db"].(string), ";", "HOSTNAME=", dashDB.Credentials["hostname"].(string), ";",
    "PORT=", strconv.FormatFloat(dashDB.Credentials["port"].(float64), 'f', 0, 64), ";", "PROTOCOL=TCPIP", ";", "UID=", dashdbuser, ";", "PWD=", dashdbpass}
  conn := strings.Join(connStr, "")

  db, err := sqlx.Connect("db2-cli", conn)
  if err != nil {
    ReturnError(w, 500, "db2_error", "Unable to open connection.")
    return
  }
  db.MapperFunc(strings.ToUpper)
  defer db.Close()


  var totalNumRows int64

  row := db.QueryRow(`SELECT count(*) FROM SETIUSERS.ACA_CANDIDATE_COORDINATES 
    WHERE RA2000HR > ? AND RA2000HR < ? AND DEC2000DEG > ? AND DEC2000DEG < ?`,ramin, ramax, decmin, decmax)

  err = row.Scan(&totalNumRows)
  if err != nil {
    ReturnError(w, 500, "query_count_error", err.Error())
    return
  }

  knownACACoordinates := []KnownACACoordinate{}
  err = db.Select(&knownACACoordinates, `SELECT * FROM SETIUSERS.ACA_CANDIDATE_COORDINATES 
    WHERE RA2000HR > ? AND RA2000HR < ? AND DEC2000DEG > ? AND DEC2000DEG < ? 
    ORDER BY RA2000HR, DEC2000DEG LIMIT ? OFFSET ?`,ramin, ramax, decmin, decmax, limit,skiprows)
  
  if err != nil {
    ReturnError(w, 500, "query_rows_error", err.Error())
    return
  }

  type ReturnData struct {
    TotalNumRows int64 `json:"total_num_rows"`
    Skip int64 `json:"skipped_num_rows"`
    Size int `json:"returned_num_rows"`
    Data []KnownACACoordinate `json:"rows"`
  }

  returnData := ReturnData{TotalNumRows: totalNumRows, Skip:skiprows, Size:len(knownACACoordinates), Data:knownACACoordinates}

  w.Header().Set("Content-Type", "application/json; charset=UTF-8")
  w.WriteHeader(http.StatusOK)
  if err := json.NewEncoder(w).Encode(returnData); err != nil {
    panic(err)
  }
}

