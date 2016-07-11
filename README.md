# SETI Public Data Server

## API Reference

### Endpoints:

  * [**/v1/coordinates/aca**](#celestial-coordinates-of-candidate-events)
  * [**/v1/aca/meta/{ra}/{dec}**](#meta-data-and-location-of-candidate-events)
  * [**/v1/token/{username}/{email address}**](#token-for-raw-data-access)
  * [**/v1/data/url/{container}/{objectname}**](#temporary-url-for-raw-data)
  * [**/v1/data/raw/{container}/{objectname}**](#raw-data-to-be-deprecated)

____

### Celestial Coordinates of Candidate Events
##### GET /v1/coordinates/aca

**Description**: Returns a JSON object that lists the exact RA and DEC coordinates
for 'Candidate' observations available in the Public SETI database. 
The structure of the returned JSON object is

```
{
 "returned_num_rows": 200, 
 "skipped_num_rows": 0, 
 "rows": [
  {
   "dec2000deg": 46.997, 
   "number_of_candidates": 392, 
   "ra2000hr": 19.832
  },
  ... 
 ], 
 "total_num_rows": 1124
}
```

Each element of `rows` contains the exact coordinates and the number of candidate
events found for that position (`number_of_candidates`). 

  * **Optional Parameters**

    **ramin, ramax, decmin, decmax**: any or all of these parameters may be used. 
    They define allowed ranges to be returned. 
    [This map](http://www.hpcf.upr.edu/~abel/phl/exomaps/all_stars_with_exoplanets_white.png), 
    along with the `Observing Progress` link found on [https://setiquest.info](http://setiquest.info/),
    may be helpful.

    **skip**: number of results to skip
  
    **limit**: number of results to return (maximum is 200)

  * **Examples**:

    Get candidate events found within an area of the sky.
    
    ```
    GET /v1/coordinates/aca?ramin=19.0&ramax=20.0&decmin=35.0&decmax=55.0
    ```

    Paginate through candidate coordinates

    ```
    GET /v1/coordinates/aca?skip=200
    GET /v1/coordinates/aca?skip=400
    ... until `returned_num_rows` == 0
    ```



### Meta-data and location of Candidate Events
##### GET /v1/aca/meta/{ra}/{dec}

**Description**: Given the RA and DEC celestial coordinates for a position in the sky,
returns a JSON object containing the meta-data and file location of 
each candidate event for that RA/DEC coordinate. The meta-data are the data found
in the [SignalDB](https://github.com/ibmjstart/SETI/docs/signaldb.md). 
There may be tens to thousands of candidate events for a particular position.
The structure of the JSON document is

```
{
 "returned_num_rows": 200, 
 "skipped_num_rows": 0, 
 "rows": [
  {
   "inttimes": 94, 
   "pperiods": 27.36000061, 
   "pol": "mixed", 
   "tgtid": 150096, 
   "sigreason": "PsPwrT", 
   "freqmhz": 6113.461883333, 
   "dec2000deg": 46.997, 
   "container": "setiCompAmp", 
   "objectname": "2014-05-20/act14944/2014-05-20_13-00-01_UTC.act14944.dx2016.id-0.L.archive-compamp", 
   "ra2000hr": 19.832, 
   "npul": 3, 
   "acttype": null, 
   "power": 50.652, 
   "widhz": 2.778, 
   "catalog": "keplerHZ", 
   "snr": 32.397, 
   "uniqueid": "kepler8ghz_14944_2016_0_2208930", 
   "beamno": 2, 
   "sigclass": "Cand", 
   "sigtyp": "Pul", 
   "tscpeldeg": 78.558, 
   "drifthzs": -1.055, 
   "candreason": "Confrm", 
   "time": "2014-05-20T12:59:55Z", 
   "tscpazdeg": 307.339
  }, 
  ...
  "total_num_rows": 392
}
```

Only 200 rows may be returned in a single query. Use the `skip` and `limit` options to paginate through
the results. 



  * **Optional Parameters**

    **skip**: number of results to skip
  
    **limit**: number of results to return (maximum is 200)

  * **Examples**:

    Get candidate events found within an area of the sky.
    
    ```
    GET /v1/aca/meta/19.832/46.997
    ```

    Paginate through candidate coordinates

    ```
    GET /v1/aca/meta/19.832/46.997?skip=200
    GET /v1/aca/meta/19.832/46.997?skip=400
    ... until `returned_num_rows` == 0
    ```

### Token for raw data access
##### GET /v1/token/{username}/{email address}

**Description**: This is not yet implemented. Once implemented, this will create a 
token for the user to access the data. 


### Temporary URL for raw data
##### GET /v1/data/url/{container}/{objectname}

**Description**: Given a container and object name, returns a JSON object 
containing a temporary URL to the data file.
The temporary URL will be valid for 60 minutes from the time it was issued. 
The container and object name are obtained from the results of 
[`/v1/aca/meta/{ra}/{dec}'](#meta-data-and-location-of-candidate-events)

*When the `token` parameter is implemented, you will be rate-limited to 10k 
temporary URLs per month. This is equivalent to 10 GB of raw data.*

*If you can make a compelling argument, you may obtain
an increase in the number of temporary URLs per month. Contact
adamcox@us.ibm.com*



  * **Required Parameters**

    **token**: Not yet implemented. Once implemented, a token will be required for
    data to be returned.

  * **Examples**:


    ```python
    import requests

    cont = 'setiCompAmp'
    objname = '2014-05-20/act14944/2014-05-20_13-00-01_UTC.act14944.dx2016.id-0.L.archive-compamp'

    data_url = 'https://setigopublic.mybluemix.net/v1/data/url/{}/{}'.format(cont, objname)  
    r_data_url = requests.get(data_url)

    print json.dumps(r.json(), indent=1)
    ```

    Prints 

    ```
    {
     "temp_url": "https://dal05.objectstorage.softlayer.net/v1/AUTH_4f10e4df-4fb8-44ab-8931-1529a1035371/setiCompAmp/2014-05-20/act14944/2014-05-20_13-00-01_UTC.act14944.dx2016.id-0.L.archive-compamp?temp_url_sig=b228c472264ff49fc869ada18c3ac4a0cc96817d&temp_url_expires=1468279610"
    }
    ```

    Use URL to get data.

    ```python
    r_data = requests.get(r_data_url.json()['temp_url'])

    rawdata = r_data.content
    print 'file size: ', len(rawdata)
    ```

    ```
    file size:  1061928
    ```

### Raw Data (to be deprecated)
##### GET /v1/data/raw/{container}/{objectname}

**Description**: Given a container and object name, returns the data file directly.
The container and object name are obtained from the results of 
[`/v1/aca/meta/{ra}/{dec}'](#meta-data-and-location-of-candidate-events)


  * **Examples**:

  ```python
    import requests

    cont = 'setiCompAmp'
    objname = '2014-05-20/act14944/2014-05-20_13-00-01_UTC.act14944.dx2016.id-0.L.archive-compamp'

    data_url = 'https://setigopublic.mybluemix.net/v1/data/raw/{}/{}'.format(cont, objname)  
    r_data = requests.get(data_url)

    rawdata = r_data.content
    print 'file size: ', len(rawdata)
    ```

    Prints 

    ```
    file size:  1061928
    ```
