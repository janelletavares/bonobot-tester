#!/bin/bash

get_new_token() {
curl -s $API/api/collections/users/auth-with-password \
-X POST \
-H "Content-Type: application/json" \
  -H 'Accept: */*' \
  -H 'Accept-Language: en-US' \
  -d "{\"identity\": \"$API_EMAIL\", \"password\": \"$API_PASSWORD\"}"  | python3 -m json.tool
}

create_participant_request() {
curl -s $API/api/collections/requests/records \
-X POST \
-H "Content-Type: application/json" \
  -H 'Accept: */*' \
  -H 'Accept-Language: en-US' \
  -H "Authorization: $API_TOKEN" \
  -d "{\"meeting_id\": \"$1\", \"meeting_passcode\": \"$2\", \"duration\": \"$3\", \"count\": \"$4\"}"  | python3 -m json.tool
}

list_participant_requests() {
curl -s $API/api/collections/requests/records \
-X GET \
-H "Content-Type: application/json" \
  -H 'Accept: */*' \
  -H 'Accept-Language: en-US' \
  -H "Authorization: $API_TOKEN" | python3 -m json.tool
}
