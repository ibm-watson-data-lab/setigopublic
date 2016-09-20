package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	cfenv "github.com/cloudfoundry-community/go-cfenv"

	"errors"
	"encoding/json"
	_ "bitbucket.org/phiggins/db2cli"
	"github.com/jmoiron/sqlx"
	"strings"
)

const (
	DEFAULT_PORT = "8080"
	APP_VERSION  = "0.0.2"
)


var (
	dbConnection *sqlx.DB
)

var (
	dashDB     cfenv.Service
	dashdbuser string
	dashdbpass string
)

func getDashDBCreds() (cfenv.Service, string, string) {

	// Returns the dashdB Service, plus DASHDBUSER, DASHDBPASS

	//get dashDB service details
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


func main() {

	var port string
	fmt.Println("Searching for PORT")

	appEnv, err := cfenv.Current()
	if err != nil {
		//we are not in a CF environment. Attempt to get PORT from local envars
		if port = os.Getenv("PORT"); len(port) == 0 {
			port = DEFAULT_PORT
		}
	} else {
		port = strconv.Itoa(appEnv.Port)
	}

	oauthConfig.RedirectURL = os.Getenv("OATH_REDIRECT_URL") + "/auth"

	dashDB, dashdbuser, dashdbpass = getDashDBCreds()

	connStr := []string{"DATABASE=", dashDB.Credentials["db"].(string), ";", "HOSTNAME=", dashDB.Credentials["hostname"].(string), ";",
		"PORT=", strconv.FormatFloat(dashDB.Credentials["port"].(float64), 'f', 0, 64), ";", "PROTOCOL=TCPIP", ";", "UID=", dashdbuser, ";", "PWD=", dashdbpass}
	conn := strings.Join(connStr, "")

	dbConnection, err = sqlx.Connect("db2-cli", conn)
	if err != nil {
		panic(errors.New("Failed to connected to DashDB."))
	}
	dbConnection.MapperFunc(strings.ToUpper)

	router := NewRouter()

	fmt.Println("listening on port " + port + " ...")

	log.Fatal(http.ListenAndServe(":"+port, router))

}
