# conn-checker
Tool to batch check and validate connections to URL's.

### SPEC
1. http url validation service
    - interface: 1-2 validate-url endpoints e.g. a single url and batch
    - in: json list of urls
    - out: some meaningful collection of url validations based
        - collect successful http status codes
        - collect errors and suggestions
        - return as json

Example in/output:
```
// in
{
	"urls": [{
			"id": 0,
			"url": "https://example.com"
		},
		{
			"id": 1,
			"url": "http://www.somewhere.com"
		}
	]
}

// out
{
	"validations": {
		"valid": [{
			"id": 1,
			"req_url": "http://www.somewhere.com",
			"end_url": "https://www.somewhere.com/",
			"status": 200
		}],

		"errors": [{
			"id": 0,
			"req_url": "https://example.com",
			"end_url": "https://example.com",
			"status": 404,
			"suggestion": "Do this and that and then try again."
		}]
	}
}
```

### TODO
- keep intermediate results in memory
- rewrite to use json data instead of csv
- add api endpoints


#### flow
request -> 
    json to UrlJob -> 
        add to work queue ->
            collect results in two piles as json -> 
                return to sender