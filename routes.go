package main

import (
    "net/http"
)

type Route struct {
    Name        string
    Method      string
    Pattern     string
    HandlerFunc http.HandlerFunc
}
type Routes []Route


var routes = Routes{
    Route{
        "Index",
        "GET",
        "/",
        Index,
    },
    Route{
        "ACASingleFileByCoordinates",
        "GET",
        "/v1/aca/single", //v1/aca/single/{ra,dec} ??
        AcaByCoordinates,
    },
    // Route{
    //     "AcaBlockByTGTID",
    //     "GET",
    //     "/v1/aca/block/{tgtid}
    //     AcaBlockByTgtid,
    // },
    Route{
        "KnownCandCoordinates",
        "GET",
        "/v1/coordinates/aca",
        KnownCandCoordinates,
    },
    // ### Potential future API ### 
    // Route{
    //     "CompampSingleFileByCoordinates",
    //     "GET",
    //     "/v1/compamp/rawfile", //v1/compamp/rawfile/{ra,dec} ??
    //     CompampSingleFileByCoordinates,
    // }
    // Route{
    //     "AllKnownCoordinates",
    //     "GET",
    //     "/v1/coordinates",  //with many options
    //     AllKnownCoordinates,
    // },
    // Route
    //     "AcaByCoordinates",
    //     "GET",
    //     "/v1/kepler_target",
    //     AcaByCoordinates,
    // },
  }