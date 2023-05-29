# MokAPI

Simple mocking API to return JSON responses. 

Service is looking over the given folder if any changes occur on JSON file it will reload.
Watch interval can be specified as argument.

# Definition structure

To define an endpoint in JSON format use this template
```json
{
  "endpoint": "/<path>",
  "method": "<http_method>",
  "response_status_code": <status_code>,
  "response_payload": {
    <valid_json>
  }
}
```
Response payload isn't parsed the only criteria is to be a valid JSON.

Save your structure in JSON files in one folder that will be used as run argument.

# Running locally

For running locally you will need GO on your machine and folder with all your endpoint definitions in JSON format.

Example of command
```shell
go run  . -definitions-path ./example_data -check-interval 2s
```

To see all the arguments run `go run . -help`

# Adding definitions using HTTP request

Service is exposing endpoint `POST:/mokapi/add` to be used for adding definitions.

Example of request
```shell
curl --location 'localhost:8080/mokapi/add' \
--header 'Content-Type: application/json' \
--data '{
  "endpoint": "/",
  "method": "DELETE",
  "response_status_code": 200,
  "response_payload": {
    "status": "OK BUT NOT DELETE"
  }
}'
```

## Known TODOs
- [ ] Release as package or GO binary
- [ ] Github Actions for releases
- [ ] Support more formats than JSON
