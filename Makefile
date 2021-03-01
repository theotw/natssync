CLOUD_OPENAPIDEF_FILE=openapi/bridge_server_v1.yaml
CLIENT_OPENAPIDEF_FILE=openapi/bridge_client_v1.yaml
openapicli_jar=third_party/openapi-generator-cli.jar

ifndef IMAGE_TAG
	IMAGE_TAG=latest
endif

ifndef IMAGE_REPO
	IMAGE_REPO=theotw
endif

ifeq (${IMAGE_TAG},latest)
	BUILD_VERSION=$(shell date '+%Y%m%d%H%M')
else
	BUILD_VERSION=${IMAGE_TAG}
endif

tmp:

	echo ${IMAGE_TAG}
	echo ${BUILD_VERSION}

generate: maketmp justgenerate rmtmp
maketmp:
	rm -r -f tmpcloud
	mkdir -p tmpcloud
	rm -r -f tmpclient
	mkdir -p tmpclient

rmtmp:
	rm -r -f tmpcloud
	rm -r -f tmpclient

echoenv:
	echo "PATH ${PATH}"
	echo "REPO ${IMAGE_REPO}"
	echo "TAG ${IMAGE_TAG}"

justgenerate: generateserver generateclient generateversion
generateserver:
	docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -g go-server --package-name v1 -i /local/${CLOUD_OPENAPIDEF_FILE} -o /local/tmpcloud
	rm -r -f pkg/bridgemodel/generated/v1
	mkdir -p pkg/bridgemodel/generated/v1
	cp tmpcloud/go/model* pkg/bridgemodel/generated/v1

generateclient:
	docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -g go-server --package-name v1 -i /local/${CLIENT_OPENAPIDEF_FILE} -o /local/tmpclient
	rm -r -f pkg/bridgeclient/generated/v1
	mkdir -p pkg/bridgeclient/generated/v1
	cp tmpclient/go/model* pkg/bridgeclient/generated/v1

generateversion:
	echo "//THIS IS A GENERATED FILE, any changes will be overridden " >pkg/version.go
	echo "package pkg" >>pkg/version.go
	echo "const VERSION=\"${BUILD_VERSION}\"" >>pkg/version.go

incontainergenerate:generateversion
	rm -r -f tmpcloud
	rm -r -f tmpclient
	mkdir tmpcloud
	mkdir tmpclient

	java -jar ${openapicli_jar} generate -g go-server --package-name v1 -i ${CLOUD_OPENAPIDEF_FILE} -Dmodels -o tmpcloud
	java -jar ${openapicli_jar} generate -g go-server --package-name v1 -i ${CLIENT_OPENAPIDEF_FILE} -Dmodels -o tmpclient

	rm -r -f pkg/bridgemodel/generated
	mkdir -p pkg/bridgemodel/generated/v1
	echo "THIS IS A GENERATED DIR, DONT PUT FILES HERE" >pkg/bridgemodel/generated/readme.txt
	cp tmpcloud/go/*  pkg/bridgemodel/generated/v1

	rm -r -f pkg/bridgeclient/generated/v1
	mkdir -p pkg/bridgeclient/generated/v1
	echo "THIS IS A GENERATED DIR, DONT PUT FILES HERE" >pkg/bridgeclient/generated/readme.txt
	cp tmpclient/go/model* pkg/bridgeclient/generated/v1

	rm -rf tmpcloud
	rm -rf tmpclient

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

buildarm: export GOOS=linux
buildarm: export GOARCH=arm
buildarm: export CGO_ENABLED=0
buildarm: export GO111MODULE=on
buildarm:
	mkdir -p out
	rm -f  out/bridgeserver_x64_linuxarm64
	go build -v -o out/bridgeserver_x86_linux_arm apps/bridge_server.go
	go build -v -o out/bridgeclient_x86_linux_arm apps/bridge_client.go
	go build -v -o out/echo_main_x86_linux_arm apps/echo_main.go
	go build -v -o out/echo_client_x86_linux_arm apps/echo_client.go
	go build -v -o out/simple_auth_x86_linux_arm apps/simple_reg_auth_server.go

clean:
	rm -r -f tmp
	rm -r -f pkg/bridgemodel/generated/v1
	rm -r -f out
	rm go.sum


cloudimage:
	docker build --no-cache --build-arg IMAGE_REPO=${IMAGE_REPO} --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/natssync-server:${IMAGE_TAG} --target natssync-server .
cloudimageBuildAndPush:cloudimage
	docker push ${IMAGE_REPO}/natssync-server:${IMAGE_TAG}
cloudimagearm:
	docker build --no-cache --build-arg IMAGE_REPO=${IMAGE_REPO} --build-arg IMAGE_TAG=${IMAGE_TAG} -f DockerfileArm --tag ${IMAGE_REPO}/natssync-server-arm:${IMAGE_TAG} --target natssync-server-arm .

debugcloudimage:
	docker build --no-cache --build-arg IMAGE_REPO=${IMAGE_REPO} --build-arg IMAGE_TAG=${IMAGE_TAG} -f CloudServerDebug.dockerfile --tag ${IMAGE_REPO}/debugnatssync-server:${IMAGE_TAG} .


testimage:
	docker build --no-cache --build-arg IMAGE_REPO=${IMAGE_REPO} --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/natssync-tests:${IMAGE_TAG} --target natssync-tests .
testimageBuildAndPush: testimage
	docker push ${IMAGE_REPO}/natssync-tests:${IMAGE_TAG}

clientimage:
	docker build --no-cache --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/natssync-client:${IMAGE_TAG} --target natssync-client .
clientimageBuildAndPush: clientimage
	docker push ${IMAGE_REPO}/natssync-client:${IMAGE_TAG}
clientimagearm:
	docker build --no-cache -f DockerfileArm --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/natssync-client-arm:${IMAGE_TAG} --target natssync-client-arm .

echoproxylet:
	docker build --no-cache --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/echo-proxylet:${IMAGE_TAG} --target echo-proxylet .
echoproxyletBuildAndPush: echoproxylet
	docker push ${IMAGE_REPO}/echo-proxylet:${IMAGE_TAG}

echoproxyletarm:
	docker build --no-cache -f DockerfileArm --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/echo-proxylet-arm:${IMAGE_TAG} --target echo-proxylet-arm .

simpleauth:
	docker build --no-cache --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/simple-reg-auth:${IMAGE_TAG} --target simple-reg-auth .
simpleautharm:
	docker build --no-cache -f DockerfileArm --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/simple-reg-auth-arm:${IMAGE_TAG} --target simple-reg-auth-arm .

simpleauthBuildAndPush: simpleauth
	docker push ${IMAGE_REPO}/simple-reg-auth:${IMAGE_TAG}

allimages:
	docker-compose -f docker-compose.yml build
allarmimages:
	docker-compose -f docker-compose-arm.yml -f DockerfileArm build

allimagesBuildAndPush:testimageBuildAndPush cloudimageBuildAndPush clientimageBuildAndPush echoproxyletBuildAndPush simpleauthBuildAndPush


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
