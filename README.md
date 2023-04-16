```
                                 _               _
                                | |             | |
  ___ ___  _ __  _ __ ______ ___| |__   ___  ___| | _____ _ __
 / __/ _ \| '_ \| '_ \______/ __| '_ \ / _ \/ __| |/ / _ \ '__|
| (_| (_) | | | | | | |    | (__| | | |  __/ (__|   <  __/ |
 \___\___/|_| |_|_| |_|     \___|_| |_|\___|\___|_|\_\___|_|

Tool to batch check and validate connections to URL's.
```

## CLI
To work with .csv format from the command line.

```
Description
        Reads a bunch of URL end points from a csv file and returns another csv file containing the HTTP
        and robotstxt results from contacting each endpoint. The input should be formatted in rows of
        {id, url}. The URL does not need to be well formatted. Currently output is stored on disk
        incrementally as it is collected. Output is found in the ./output/ folder.

Usage
  -file string
        Path to the input .csv file.
```

The expected .csv format:
```csv
id,url
1,url1.com
2,url2.edu
3,www.url3.com
4,https://www.url3.com
```
Note id's do not have to be numeric nor sequential.

Example to process a `.csv` file from the project root and store results in `output/`:
```bash
go run cmd/cli/main.go -file csv_file
```

## HTTP API
To work with json based requests.

Example curl request:
```
 curl -X POST localhost:8080/validate \
 	-H 'Content-Type: application/json' \
 	-d '[{"id":"0","url":"www.example.com"}]'
```

Example json request:
```json
[
  {
      "id": "0",
      "req_url": "blablainvalidurl.coma"
  },
  {
    "id": "1",
    "url": "fineurl1.com"
  },
  {
    "id": "2",
    "url": "fineurl2.com"
  }
]
```

Example json response:
```json
{
  "http_success": [
    {
      "id": "1",
      "end_url": "fineurl1.com",
      "http_status": 200,
      "robots_ok": true
    },
    {
      "id": "2",
      "end_url": "fineurl2.com",
      "http_status": 200,
      "robots_ok": true
    }
  ],
  "http_errors": [
    {
      "id": "0",
      "req_url": "blablainvalidurl.coma",
      "suggestion": "Get \"http://blablainvalidurl.coma\": dial tcp: lookup blablainvalidurl.coma: no such host"
    }
  ],
  "other_errors": []
}
```
