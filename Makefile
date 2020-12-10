CLOUD_OPENAPIDEF_FILE=openapi/cloud_openapi_v1.yaml

generate:
	rm -r -f tmpcloud
	mkdir tmpcloud
	docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -g go-server --package-name v1 -i /local/${CLOUD_OPENAPIDEF_FILE} -o /local/tmpcloud
	rm -f -f pkg/bridgemodel/generated/v1
	mkdir pkg/bridgemodel/generated/v1
	cp tmpcloud/go/model* pkg/bridgemodel/generated/v1
	rm -r -f tmpcloud
