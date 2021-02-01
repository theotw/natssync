CLOUD_OPENAPIDEF_FILE=openapi/cloud_openapi_v1.yaml
openapicli_jar=third_party/openapi-generator-cli.jar


ifndef IMAGE_TAG
	IMAGE_TAG=latest
endif
ifndef IMAGE_REPO
	IMAGE_REPO=theotw
endif



generate: maketmp justgenerate rmtmp
maketmp:
	rm -r -f tmpcloud
	mkdir -p tmpcloud

rmtmp:
	rm -r -f tmpcloud

echoenv:
	echo "Version 1"
	echo "PATH ${PATH}"
	echo "REPO ${IMAGE_REPO}"
	echo "TAG ${IMAGE_TAG}"
justgenerate:
	docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -g go-server --package-name v1 -i /local/${CLOUD_OPENAPIDEF_FILE} -o /local/tmpcloud
	rm -r -f pkg/bridgemodel/generated/v1
	mkdir pkg/bridgemodel/generated/v1
	cp tmpcloud/go/model* pkg/bridgemodel/generated/v1

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
	go build -v -o out/simple_auth_x64_darwin apps/simple_reg_auth_server.go


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
	go build -v -o out/simple_auth_x64_linux apps/simple_reg_auth_server.go

clean:
	rm -r -f tmp
	rm -r -f pkg/bridgemodel/generated/v1
	rm -r -f out
	rm go.sum


cloudimage:
	docker build --no-cache --build-arg IMAGE_REPO=${IMAGE_REPO} --build-arg IMAGE_TAG=${IMAGE_TAG} -f CloudServer.dockerfile --tag ${IMAGE_REPO}/natssync-server:${IMAGE_TAG} .
cloudimageBuildAndPush:cloudimage
	docker push ${IMAGE_REPO}/natssync-server:${IMAGE_TAG}

testimage:
	docker build --no-cache -f NatssyncTestImage.dockerfile  --tag ${IMAGE_REPO}/natssync-tests:${IMAGE_TAG} .
testimageBuildAndPush: testimage
	docker push ${IMAGE_REPO}/natssync-tests:${IMAGE_TAG}

clientimage:
	docker build --no-cache -f CloudClient.dockerfile --tag ${IMAGE_REPO}/natssync-client:${IMAGE_TAG} .
clientimageBuildAndPush: clientimage
	docker push ${IMAGE_REPO}/natssync-client:${IMAGE_TAG}

echoproxylet:
	docker build --no-cache -f EchoProxylet.dockerfile --tag ${IMAGE_REPO}/echo-proxylet:${IMAGE_TAG} .
echoproxyletBuildAndPush: echoproxylet
	docker push ${IMAGE_REPO}/echo-proxylet:${IMAGE_TAG}

simpleauth:
	docker build --no-cache -f SimpleAuthServer.dockerfile --tag ${IMAGE_REPO}/simple-reg-auth:${IMAGE_TAG} .

simpleauthBuildAndPush: simpleauth
	docker push ${IMAGE_REPO}/simple-reg-auth:${IMAGE_TAG}

allimages: cloudimage clientimage echoproxylet simpleauth testimage

allimagesBuildAndPush: cloudimageBuildAndPush clientimageBuildAndPush testimageBuildAndPush echoproxyletBuildAndPush simpleauthBuildAndPush


pushall:
	docker push ${IMAGE_REPO}/natssync-server:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/natssync-client:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/echo-proxylet:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/simple-reg-auth:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/natssync-tests:${IMAGE_TAG}


l1:
	echo ${PATH}
	go get github.com/jstemmer/go-junit-report
	mkdir -p out
	#go test -v -coverpkg=github.com/theotw/natssync/pkg/... -coverprofile=out/unit_coverage.out github.com/theotw/natssync/pkg/...
	go test -v -coverpkg=github.com/theotw/natssync/pkg/... -coverprofile=out/unit_coverage.out github.com/theotw/natssync/pkg/...  2>&1 >out/l1_out.txt
	cat out/l1_out.txt | go-junit-report > out/report_l1.xml
