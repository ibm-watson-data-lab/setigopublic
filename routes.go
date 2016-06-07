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
        "/v1/aca/meta/single", //v1/aca/meta/single/{ra,dec} ??
        AcaByCoordinates,
    },
    // Route{
    //     "AcaBlockByTGTID",
    //     "GET",
    //     "/v1/aca/meta/block/{tgtid}
    //     AcaBlockByTgtid,
    // },
    Route{
        "KnownCandCoordinates",
        "GET",
        "/v1/coordinates/aca",
        KnownCandCoordinates,
    },
//        "/v1/aca/url/{container}/{objectname:\"[a-zA-Z0-9=\\-\\/.]+}\"},  //this regex doesn't work!
    Route{
        "ACAURL",
        "GET",
        "/v1/aca/url/{container}/{date}/{act}/{object}",
        GetACARawDataTempURL,
    },
        Route{
        "ACAData",
        "GET",
        "/v1/aca/data/{container}/{date}/{act}/{object}",
        GetACARawData,
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