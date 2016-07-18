# SETI Public Data Server

Go-based server for Public SETI data sets, part of the IBM+SETI partnership. 

The Allen Telescope Array (ATA), the source of data for this project,
records radio signals from particular celestial coordinates 
through beamforming.
Beamforming combines signals from multiple telescopes in order to observe radio signals from just a 
very small region of the sky, usually occupied by a single star with a known exoplanet.

Observation of a potential signal in a beam results in two data files: a raw data file, called a `compamp`
or `archive-compamp` and preliminary measurement and categorization of the signal, which is stored in a row
of the SignalDB. The SignalDB is a single table, managed with 
[dashDB](http://www.ibm.com/analytics/us/en/technology/cloud-data-services/dashdb), that contains the 
preliminary analysis of the raw data. A SignalDB row contains signal categorization, the Right Ascension (RA)
and Declination (DEC) coordinates of the position in the sky, estimates of the power of the signal, 
primary carrier frequency, etc. All RA/DEC coordinates are references from the J2000 equinox. A full
description of the values in SignalDB are found [here]().

The raw data for a signal that is categorized as a `Candidate` is stored as an `archive-compamp` file, 
while other types signals are stored as `compamp` files. The only difference between these file types
is that the `archive-compamp` contains all the data across the entire bandwith observed by the ATA, while
the `compamp` files contains just the data for the subband where the signal was observed. The full bandwidth
is divided into 16 subbands.  An `archive-compamp` file is typically ~1 MB in size. As such, a typical
`compamp` file is ~1/16 MB. 

This REST API allows you to search for `Candidate`/`archive-compamp` files based upon their position
in the sky, the RA/DEC values. Further enhancements to this API will allow for finding data based
on other attributes.

## Full Example Usage with Python

### Introduction

The following example, done in Python, shows a typical way to use this REST API. 

The general flow is

  * Determine if data is available for a particular point or region in the sky.
  * Get the SignalDB rows and raw data file information for that region. 
  * Get a temporary URL for the raw data file.
  * Download the data and store it.

Though it's not necessary, typically one will start by selecting an interesting target or
interesting region in the sky.  

Once the authorization token system is implemented, you will be required to have an [IBM Bluemix](https://bluemix.net)
account and will be limited to 10k temporary URL requests per month. 

You may read below or [click here to see a full example in a shared IBM Spark notebook](https://console.ng.bluemix.net/data/notebooks/e17dc8c6-9c33-4947-be31-ee6b4b7e0888/view?access_token=6e95d320610f67467ba63bc89d9cec48faf847f2532fdd7523b0dd2ccb9ea346).

### Select an Interesting Target

Use some kind of [source](http://phl.upr.edu/projects/habitable-exoplanets-catalog)
to find interesting expolanet coordinates. There are other websites that carry this information.
You can also pick a region in the sky. 
[This map](http://www.hpcf.upr.edu/~abel/phl/exomaps/all_stars_with_exoplanets_white.png), 
along with the `Observing Progress` link found on [https://setiquest.info](http://setiquest.info/),
may be helpful.

You'll need to know the right ascension (RA, in hours from 0.0 to 24) and declination 
(DEC, in degrees from -90.0 to 90.0) of your exoplanet/target of interest. However, the RA/DEC
values recorded in the database may not exactly match the known position of the object (at most, 
these values differ by 0.002 degrees). Or you may choose a wider region of the sky. Either way, 
we use the API to search for data within an enclosed region of the sky. 

#### Kepler 1229b

Let's look at Kepler 1229b, found here http://phl.upr.edu/projects/habitable-exoplanets-catalog.
This candidate planet has an Earth similarity index (ESI) of 0.73 (Earth = 1.0, Mars = 0.64, Jupiter = 0.12) 
and is 770 light-years away. It is the 5th highest-ranked planet by ESI.

You can take a look at this object in the sky: 
[http://simbad.cfa.harvard.edu/simbad/sim-id?Ident=%408996422&Name=KOI-2418.01&submit=submit](http://simbad.cfa.harvard.edu/simbad/sim-id?Ident=%408996422&Name=KOI-2418.01&submit=submit)

The celestial coordinates for Kepler 1229b is RA = 19.832 and DEC = 46.997

### Query the SignalDB

Here is the example code to query the API for a region about the position of Kepler 1229b. The API
endpoint to use is [`/v1/coordinates/aca`](#celestial-coordinates-of-candidate-events). In this case,
we define a 0.004 x 0.004 box around the RA/DEC position.

```python
import requests
RA=19.832
DEC=46.997
box = 0.002

# We want this query
# http://setigopublic.mybluemix.net/v1/coordinates/aca?ramin=19.830&ramax=19.834&decmin=46.995&decmax=46.999

params = {
  'ramin':RA-box, 'ramax':RA+box, 'decmin':DEC-box, 'decmax':DEC+box
}
r = requests.get('https://setigopublic.mybluemix.net/v1/coordinates/aca',
    params = params)

import json
print json.dumps(r.json(), indent=1)
```

We get the following output:

```json
{
  "returned_num_rows": 1,   
  "skipped_num_rows": 0, 
  "rows": [
    {
      "dec2000deg": 46.997, 
      "number_of_rows": 392, 
      "ra2000hr": 19.832
    }
  ], 
  "total_num_rows": 1
}
```

This means we found 392 'Candidate' signals recorded while observing this planet in the SignalDB. 

##### Aside

If we expand our range of allowed RA/DEC values, we find 'Candidate' events for some nearby positions.

```
GET http://setigopublic.mybluemix.net/v1/coordinates/aca?ramin=19.800&ramax=19.90&decmin=46.95&decmax=47.02
```

returns

```json
{
  "total_num_rows": 4,
  "skipped_num_rows": 0,
  "returned_num_rows": 4,
  "rows": [
    {
      "ra2000hr": 19.832,
      "dec2000deg": 46.997,
      "number_of_candidates": 392
    },
    { 
      "ra2000hr": 19.834,
      "dec2000deg": 46.961,
      "number_of_candidates": 370
    },
    {
      "ra2000hr": 19.856,
      "dec2000deg": 46.968,
      "number_of_candidates": 44
    },
    {
      "ra2000hr": 19.861,
      "dec2000deg": 46.965,
      "number_of_candidates": 32
    }
  ]
}
```

You may wish to extend your box further so see what we find. 

### Get Raw Data URLs

Given a particular celestial coordinate, we can obtain all of the 'Candidate' signal meta data and, importantly,
a URL to the raw data.

The endpoint to use is 
[/v1/aca/meta/{ra}/{dec}](#meta-data-and-location-of-candidate-events).

Continuing from the example above


```python
ra = r.json()['rows'][0]['ra2000hr']  #19.832 from above query
dec = r.json()['rows'][0]['dec2000deg'] 46.997 

r = requests.get('https://setigopublic.mybluemix.net/v1/aca/meta/{}/{}'.format(ra, dec)

print json.dumps(r.json(), indent=1)
```

Here's what the output will look like:

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

The maximum return limit is 200 rows per query. The `total_num_rows` tells you there are 392 rows. To get the rest, you'll need to use the `?skip=200` optional
URL parameter. For example

```python 
r = requests.get('https://setigopublic.mybluemix.net/v1/aca/meta/{}/{}?skip=200'.format(ra, dec)
newrows = r.json()['rows']
```

Searching through these results, one thing that you'll notice is that while there are 392
'Candidate' signals found in the raw data, it doesn't mean there are 392 raw data files. There
will be duplicates. So you'll need to sort through them appropriately. If you do not and you use
the same raw data multiple times within a machine-learning algorithm (to extract a set of features, 
for example), you'll likely corrupt your results. 

You'll also notice there multiple files that have the same SignalDB meta-data, but with slightly
different file names. The antenna data are decomposed into left- and right-circularly polarized
complex data signals, which are stored in separate files; hence, the `L` and `R` components of the names.

The location of the raw data is stored in an 
[IBM/Softlayer Object Store](https://developer.ibm.com/bluemix/2015/10/20/getting-started-with-bluemix-object-storage/).  
In order to access those  data files, one must first request a temporary URL.

### Temporary URLs and Data Storage

The temporary URLs are obtained with the [`/v1/data/url/{containter}/{objectname}`](#temporary-url-for-raw-data) 
endpoint. These temporary URLs, by default, are valid for only one hour. You must consider this when obtaining
the URLs and retrieving the data. 

In each `row` returned above, along with the SignalDB data, there is the `container` and `objectname` of the
raw data file. 

While not necessary for demonstration of this API, the intelligent thing to do is to `map` the rows into
a list of tuples that just contain the `container` and `objectname` and then drop duplicates. This will reduce the number
of temporary URLs you need and the amount of storage space you'll need to store elsewhere. 

```python
rows = r.json(['rows'])
data_paths = set( map(lambda x: (x['container'], x['objectname']), rows) )
```

Next, we use the API to get the temporary URLs

```python
def get_temp_url(row):
  r = requests.get('https://setigopublic.mybluemix.net/v1/data/url/{}/{}'.format(row[0], row[1]))
  return (r.status_code, r.json()['temp_url'], row[0], row[1])

temp_urls = map(get_temp_url, data_paths)
```

For each temporary URL, we download the file and should then put it somewhere. That is not specified 
in this example. However, you can [look here for a full example](https://console.ng.bluemix.net/data/notebooks/e17dc8c6-9c33-4947-be31-ee6b4b7e0888/view?access_token=6e95d320610f67467ba63bc89d9cec48faf847f2532fdd7523b0dd2ccb9ea346)
of doing essentially what you see here with an IBM Spark Service and then placing the data in 
an [IBM Object Store](https://developer.ibm.com/bluemix/2015/10/20/getting-started-with-bluemix-object-storage/).

```python
def move_data(row):
  if row[0] == 200:  #we have a valid temporary URL
    r = requests.get(row[1])
  if r.status_code == 200
    data = r.content
    ### DO something with the data here!
    ### It's most efficient for you and for your analysis to store it somewhere relatively local

  return (r.status_code,) + row[1:]

moved_data = map(move_data, temp_urls)
```


## API Reference

### Endpoints:

  * [**/v1/coordinates/aca**](#celestial-coordinates-of-candidate-events)
  * [**/v1/aca/meta/{ra}/{dec}**](#meta-data-and-location-of-candidate-events)
  * [**/v1/token/{username}/{email address}**](#token-for-raw-data-access)
  * [**/v1/data/url/{container}/{objectname}**](#temporary-url-for-raw-data)

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

