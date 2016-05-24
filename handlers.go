package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    _ "bitbucket.org/phiggins/db2cli"
    cfenv "github.com/cloudfoundry-community/go-cfenv"
    //"github.com/gorilla/mux"
    "database/sql"
    "html"
    "time"
    "strings"
    "strconv"
)

func Index(w http.ResponseWriter, r *http.Request) {  
  fmt.Fprintln(w, "Hello World!! %q", html.EscapeString(r.URL.Path))
}

func RowIndex(w http.ResponseWriter, r *http.Request) {
  
  rows := SignalDbRows{
    SignalDbRow{
      UniqueID: "kepler1", 
      Time: time.Now(),
      ActTyp: "target",
      TGTID: 100,
      Catalog: "kepler",
      Ra2000Hr: 20.2,
      Dec2000Deg: 13.3,
      Power: 34.4,
      SNR: 23.2,
      FreqMHZ: 6667.3323,
      DriftHZS: 3.4423,
      WIDHZ: 1.234,
      POL: "both", 
      SigType: "CwP",
      PPeriods: 11.2,
      NPul: 2,
      IntTimes: 1.2,
      TSCPAZDEG: 24.2314,
      TSCPELDEG: -12.23,
      BeamNo: 3,
      SigClass: "Cand",
      SigReason: "PsPwrt",
      CandReason: "ZeroDft"},
  }

  w.Header().Set("Content-Type", "application/json; charset=UTF-8") 
  w.WriteHeader(http.StatusOK)
  if err := json.NewEncoder(w).Encode(rows); err != nil {
        panic(err)
    }
}

func connect(w http.ResponseWriter, req *http.Request) {
  
  fmt.Fprintln(w, "<html><head><title>Sample go dashDB application</title></head><body>")
  fmt.Printf("<html><head><title>Sample go dashDB application</title></head><body>")
  
  // Fetch dashDB service details
  appEnv, _ := cfenv.Current()
  services, error := appEnv.Services.WithLabel("dashDB")
  
  if( error != nil ) {
    fmt.Fprintln(w, "<h4>No dashDB service instance bound to the applicaiton. Please bind a dashDB service instance and retry.</h4>")
    fmt.Printf("<h4>No dashDB service instance bound to the applicaiton. Please bind a dashDB service instance and retry.</h4>")
  } else {
    dashDB := services[0]
    
    connStr := []string{"DATABASE=", dashDB.Credentials["db"].(string), ";", "HOSTNAME=", dashDB.Credentials["hostname"].(string), ";", 
        "PORT=",strconv.FormatFloat(dashDB.Credentials["port"].(float64), 'f', 0, 64), ";", "PROTOCOL=TCPIP", ";", "UID=", dashDB.Credentials["username"].(string), ";", "PWD=", dashDB.Credentials["password"].(string)};
    conn := strings.Join(connStr, "")
    
    fmt.Printf(conn)

    db, err := sql.Open("db2-cli", conn)
    if err != nil {

      fmt.Fprintln(w, "<h3>" )
      fmt.Fprintln(w, err )
      fmt.Fprintln(w, "</h3>" )
      return
    }
    
    defer db.Close()
    
    var (
      first_name string
      last_name string
    )
    
    stmt, err := db.Prepare("SELECT FIRST_NAME, LAST_NAME from GOSALESHR.employee FETCH FIRST 10 ROWS ONLY")
    if err != nil {
      fmt.Fprintln(w, "<h3>" )
      fmt.Fprintln(w, err )
      fmt.Fprintln(w, "</h3>" )
      return
    }
    defer stmt.Close()
    
    rows, err := stmt.Query()
    if err != nil {
      fmt.Fprintln(w, "<h3>" )
      fmt.Fprintln(w, err )
      fmt.Fprintln(w, "</h3>" )
      return
    }
    defer rows.Close()
    fmt.Fprintln(w, "<h3>Query: <br>SELECT FIRST_NAME, LAST_NAME from GOSALESHR.employee FETCH FIRST 10 ROWS ONLY</h3>" )
    fmt.Fprintln(w, "<h3>Result:<br></h3><table border=\"1\"><tr><th>First Name</th><th>Last Name</th></tr>" )
    for rows.Next() {
      err := rows.Scan(&first_name, &last_name)
      if err != nil {
        fmt.Fprintln(w, "<h3>" )
        fmt.Fprintln(w, err )
        fmt.Fprintln(w, "</h3>" )
        return;
        
      }
      fmt.Fprintln(w, "<tr><td>", first_name, "</td><td>",last_name, "</td></tr>")
    }
  }
  fmt.Fprintln(w, "</table></body></html>")
}