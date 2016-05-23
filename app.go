package main


import (
    "fmt"
    "net/http"
    "os"
//    "html/template"
    "database/sql"
    "strings"
		"log"
    _ "bitbucket.org/phiggins/db2cli"
		cfenv "github.com/cloudfoundry-community/go-cfenv"
    "strconv"
)

// const (
// 	DEFAULT_PORT = "8080"
// )

// var index = template.Must(template.ParseFiles(
//   "templates/_base.html",
//   "templates/index.html",
// ))

// func helloworld(w http.ResponseWriter, req *http.Request) {
//   index.Execute(w, nil)
// }

// func main() {
// 	var port string
// 	if port = os.Getenv("PORT"); len(port) == 0 {
// 		port = DEFAULT_PORT
// 	}

// 	http.HandleFunc("/", helloworld)
// 	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	
// 	log.Printf("Starting app on port %+v\n", port)
// 	http.ListenAndServe(":"+port, nil)
// }

func main() {

    http.HandleFunc("/", hello)
    http.HandleFunc("/connect", connect)
    fmt.Println("listening on port "+os.Getenv("PORT")+" ...")

    err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
    if err != nil {
      //panic(err)
    log.Fatal(err);
    fmt.Println("Error processing request ", err, " ...")
    }
}

func hello(res http.ResponseWriter, req *http.Request) {
  
      fmt.Fprintln(res, "Hello World!!")
}

func connect(res http.ResponseWriter, req *http.Request) {
  
  fmt.Fprintln(res, "<html><head><title>Sample go dashDB application</title></head><body>")
  fmt.Printf("<html><head><title>Sample go dashDB application</title></head><body>")
  
  // Fetch dashDB service details
  appEnv, _ := cfenv.Current()
  services, error := appEnv.Services.WithLabel("dashDB")
  
  if( error != nil ) {
    fmt.Fprintln(res, "<h4>No dashDB service instance bound to the applicaiton. Please bind a dashDB service instance and retry.</h4>")
    fmt.Printf("<h4>No dashDB service instance bound to the applicaiton. Please bind a dashDB service instance and retry.</h4>")
  } else {
    dashDB := services[0]
    
    connStr := []string{"DATABASE=", dashDB.Credentials["db"].(string), ";", "HOSTNAME=", dashDB.Credentials["hostname"].(string), ";", 
        "PORT=",strconv.FormatFloat(dashDB.Credentials["port"].(float64), 'f', 0, 64), ";", "PROTOCOL=TCPIP", ";", "UID=", dashDB.Credentials["username"].(string), ";", "PWD=", dashDB.Credentials["password"].(string)};
    conn := strings.Join(connStr, "")
    
    fmt.Printf(conn)

    db, err := sql.Open("db2-cli", conn)
    if err != nil {

      fmt.Fprintln(res, "<h3>" )
      fmt.Fprintln(res, err )
      fmt.Fprintln(res, "</h3>" )
      return
    }
    
    defer db.Close()
    
    var (
      first_name string
      last_name string
    )
    
    stmt, err := db.Prepare("SELECT FIRST_NAME, LAST_NAME from GOSALESHR.employee FETCH FIRST 10 ROWS ONLY")
    if err != nil {
      fmt.Fprintln(res, "<h3>" )
      fmt.Fprintln(res, err )
      fmt.Fprintln(res, "</h3>" )
      return
    }
    defer stmt.Close()
    
    rows, err := stmt.Query()
    if err != nil {
      fmt.Fprintln(res, "<h3>" )
      fmt.Fprintln(res, err )
      fmt.Fprintln(res, "</h3>" )
      return
    }
    defer rows.Close()
    fmt.Fprintln(res, "<h3>Query: <br>SELECT FIRST_NAME, LAST_NAME from GOSALESHR.employee FETCH FIRST 10 ROWS ONLY</h3>" )
    fmt.Fprintln(res, "<h3>Result:<br></h3><table border=\"1\"><tr><th>First Name</th><th>Last Name</th></tr>" )
    for rows.Next() {
      err := rows.Scan(&first_name, &last_name)
      if err != nil {
        fmt.Fprintln(res, "<h3>" )
        fmt.Fprintln(res, err )
        fmt.Fprintln(res, "</h3>" )
        return;
        
      }
      fmt.Fprintln(res, "<tr><td>", first_name, "</td><td>",last_name, "</td></tr>")
    }
  }
  fmt.Fprintln(res, "</table></body></html>")
}