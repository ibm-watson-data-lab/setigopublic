package main

import (
	"net/http"
)

type Route struct {
	Name         string
	Method       string
	Pattern      string
	HandlerFunc  http.HandlerFunc
	AuthRequired bool
	RateLimited  bool
}
type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
		false,
		false,
	},
	Route{
		"Login",
		"GET",
		"/login",
		Login,
		false,
		false,
	},
	Route{
		"Logout",
		"GET",
		"/logout",
		Logout,
		false,
		false,
	},
	Route{
		"BluemixAuthCallback",
		"GET",
		"/auth",
		BluemixAuthCallback,
		false,
		false,
	},
	Route{
		"Token",
		"GET",
		"/token",
		Token,
		true,
		false,
	},
	Route{
		"ACASingleFileByCoordinates",
		"GET",
		"/v1/aca/meta/{ra}/{dec}",
		AcaByCoordinates,
		false,
		false,
	},
	Route{
		"ACAMetaForSpacecraft",
		"GET",
		"/v1/aca/meta/spacecraft",
		SpaceCraft,
		false,
		false,
	},
	// Route{
	//     "AcaBlockByTGTID",
	//     "GET",
	//     "/v1/aca/meta/block/{tgtid}
	//     AcaBlockByTgtid,
	//     false,
	//     true,
	// },
	Route{
		"KnownCandCoordinates",
		"GET",
		"/v1/coordinates/aca",
		KnownCandCoordinates,
		false,
		false,
	},
	//        "/v1/aca/url/{container}/{objectname:\"[a-zA-Z0-9=\\-\\/.]+}\"},  //this regex doesn't work!
	Route{
		"ACAURL",
		"GET",
		"/v1/data/url/{container}/{date}/{act}/{acafile}",
		GetACARawDataTempURL,
		false,
		true,
	},
	Route{
		"signaldb_aca",
		"GET",
		"/v1/aca/meta/all",
		GetSignalDBJoinedACACandidateEvents,
		false,
		false,
	},
	// ### Potential future API ###
	// Route{
	//     "CompampSingleFileByCoordinates",
	//     "GET",
	//     "/v1/compamp/rawfile", //v1/compamp/rawfile/{ra,dec} ??
	//     CompampSingleFileByCoordinates,
	//     false,
	//     true,
	// }
	// Route{
	//     "AllKnownCoordinates",
	//     "GET",
	//     "/v1/coordinates",  //with many options
	//     AllKnownCoordinates,
	//     false,
	//     true,
	// },
	// Route
	//     "AcaByCoordinates",
	//     "GET",
	//     "/v1/kepler_target",
	//     AcaByCoordinates,
	//     false,
	//     true,
	// },
}
