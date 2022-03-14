# conn-checker
Tool to batch check and validate connections to URL's.

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
