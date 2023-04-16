#!/usr/bin/env bash

# echo "Test 1 - Send invalid request"
# curl -X POST localhost:8080/validate \
# 	-H 'Content-Type: application/json' \
# 	-d '{"login":"my_login","password":"my_password"}'


# echo "Test 2 - Send valid request"
# curl -X POST localhost:8080/validate \
# 	-H 'Content-Type: application/json' \
# 	-d '[{"id":"0","url":"www.example.com"}]'

# echo "Test 3 - Send valid request w. invalid URL"
# curl -X POST localhost:8080/validate \
# 	-H 'Content-Type: application/json' \
# 	-d '[{"id":"0","url":"yadadwww.asdfghj.com"}]'

# echo "Test 4 - Send multiple valid requests in single json list"
curl -X POST localhost:8080/validate \
	-H 'Content-Type: application/json' \
	-d '@./data/very_small.json'
