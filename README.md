# conn-checker
Tool to batch check and validate connections to URL's.

### spec
1. http url validation service
    - interface: 1-2 validate-url endpoints e.g. a single url and batch
    - in: list of urls
    - out: some meaningful collection of url validations based
        - collect successful http status codes
        - collect errors and suggestions
        - return as json

### todo
- keep intermediate results in memory
- rewrite to use json data instead of csv