# Dev notes

### Sat Mar 19 04:41:00 PM CET 2022
### Spec
1. http url validation service
    - interface: 1-2 validate-url endpoints e.g. a single url and batch
    - in: json list of urls
    - out: some meaningful collection of url validations based
        - collect successful http status codes
        - collect errors and suggestions
        - return as json

#### Flow
request -> 
    json to UrlJob -> 
        add to work queue ->
            collect results in two piles as json -> 
                errors + https codes ->
            return to sender as json

