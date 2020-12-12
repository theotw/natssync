CLOUD_OPENAPIDEF_FILE=openapi/cloud_openapi_v1.yaml
openapicli_jar=third_party/openapi-generator-cli.jar

generate:
	rm -r -f tmpcloud
	mkdir tmpcloud
	docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -g go-server --package-name v1 -i /local/${CLOUD_OPENAPIDEF_FILE} -o /local/tmpcloud
	rm -r -f pkg/bridgemodel/generated/v1
	mkdir pkg/bridgemodel/generated/v1
	cp tmpcloud/go/model* pkg/bridgemodel/generated/v1
	chmod -R 777 tmpcloud
	rm -r -f tmpcloud

incontainergenerate:
	rm -r -f tmpcloud
	mkdir tmpcloud
	java -jar ${openapicli_jar} generate -g go-server --package-name v1 -i ${CLOUD_OPENAPIDEF_FILE} -Dmodels -o tmpcloud
	rm -r -f pkg/bridgemodel/generated
	mkdir -p pkg/bridgemodel/generated/v1
	echo "THIS IS A GENERATED DIR, DONT PUT FILES HERE" >pkg/bridgemodel/generated/readme.txt
	cp tmpcloud/go/*  pkg/bridgemodel/generated/v1

	rm -rf tmpcloud

buildall: buildlinux buildmac

buildmac: export GOOS=darwin
buildmac: export GOARCH=amd64
buildmac: export CGO_ENABLED=0
buildmac: export GO111MODULE=on
buildmac: export GOPROXY=${GOPROXY_ENV}
buildmac: export GOSUM=${GOSUM_ENV}
buildmac:
	mkdir -p out
	rm -f  out/bridgeserver_x64_darwin
	go build -v -o out/bridgeserver_x64_darwin apps/bridge_server.go
	go build -v -o out/bridgeclient_x64_darwin apps/bridge_client.go
	go build -v -o out/echo_main_x64_darwin apps/echo_main.go
	go build -v -o out/echo_client_x64_darwin apps/echo_client.go
	go build -v -o out/k8s_proxylet_x64_darwin apps/k8s_proxylet.go
	go build -v -o out/k8s_pub_x64_darwin apps/k8s-pub.go
	go build -v -o out/k8sApiProxyServer_x64_darwin apps/k8sapi_server.go
	go build -v -o out/k8sApiProxylet_x64_darwin apps/k8sapi_proxylet.go


buildlinux:	export GOOS=linux
buildlinux: export GOARCH=amd64
buildlinux: export CGO_ENABLED=0
buildlinux: export GO111MODULE=on
buildlinux:
	mkdir -p out
	rm -f  out/bridgeserver_x64_linux
	go build -v -o out/bridgeserver_x64_linux apps/bridge_server.go
	go build -v -o out/bridgeclient_x64_linux apps/bridge_client.go
	go build -v -o out/echo_main_x64_linux apps/echo_main.go
	go build -v -o out/echo_client_x64_linux apps/echo_client.go
	go build -v -o out/k8s_proxylet_x64_linux apps/k8s_proxylet.go
	go build -v -o out/k8s_pub_x64_linux apps/k8s-pub.go
	go build -v -o out/k8sApiProxyServer_x64_linux apps/k8sapi_server.go
	go build -v -o out/k8sApiProxylet_x64_linux apps/k8sapi_proxylet.go

clean:
	rm -r -f tmp
	rm -r -f pkg/bridgemodel/generated/v1
	rm -r -f out
	rm go.sum


cloudimage:
	docker build -f CloudServer.dockerfile --tag bmason42/cloudbridgeserver:latest .

clientimage:
	docker build -f CloudClient.dockerfile --tag bmason42/cloudbridgeclient:latest .

echoproxylet:
	docker build -f EchoProxylet.dockerfile --tag bmason42/echoproxylet:latest .


allimages: cloudimage clientimage echoproxylet
