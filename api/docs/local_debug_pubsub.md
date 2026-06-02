# Debug of pubsub while developing on local machine

Below are some useful curls to interacte with the pub sub emulator during local development.

```sh
# Get posted from dapla-api to SUP postman
curl -X POST \
  -H 'content-type: application/json' \
  --data '{"returnImmediately": true, "maxMessages": 10}' \
  http://localhost:3004/v1/projects/dapla-local-dev/subscriptions/sup-staging-postman-incoming-subscription:pull



# Emulate answer from SUP-postman (i.e. post from SUP postman to dapla-api )
MESSAGE_ID="<id from graph api when sending message>"
PAYLOAD="{\"id\": \"$MESSAGE_ID\",\"result\": \"SUCCESSFUL\",\"timestamp\": \"2026-03-12T13:18:01\"}"
ENCODED_PAYLOAD=$(echo $PAYLOAD | base64 )

curl -X POST \
  -H 'content-type: application/json' \
  --data "{\"messages\": [{\"data\": \"$ENCODED_PAYLOAD\"}]}" \
  http://localhost:3004/v1/projects/dapla-local-dev/topics/sup-staging-postman-outgoing-topic:publish

```
