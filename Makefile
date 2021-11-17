CLOUD_OPENAPIDEF_FILE=openapi/bridge_server_v1.yaml
CLIENT_OPENAPIDEF_FILE=openapi/bridge_client_v1.yaml
openapicli_jar=third_party/openapi-generator-cli.jar
OPENAPI_IMAGE=openapitools/openapi-generator-cli:v5.2.0


ifndef IMAGE_REPO
	IMAGE_REPO=theotw
endif
BASE_VERSION := $(shell cat 'version.txt')
BUILD_DATE := $(shell date '+%Y%m%d%H%M')

BUILD_VERSION := ${BASE_VERSION}.${BUILD_DATE}
ifndef IMAGE_TAG
	IMAGE_TAG := ${BUILD_VERSION}
endif



printversion:
	echo Base: ${BASE_VERSION}
	echo Date: ${BUILD_DATE}
	echo Image: ${IMAGE_TAG}
	echo Build: ${BUILD_VERSION}

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

justgenerate: generateserver generateclient
generateserver:
	docker run --rm -v "${PWD}:/local" ${OPENAPI_IMAGE} generate -g go-server --package-name v1 -i /local/${CLOUD_OPENAPIDEF_FILE} -o /local/tmpcloud
	rm -r -f pkg/bridgemodel/generated/v1
	mkdir -p pkg/bridgemodel/generated/v1
	cp tmpcloud/go/model* pkg/bridgemodel/generated/v1

generateclient:
	docker run --rm -v "${PWD}:/local" ${OPENAPI_IMAGE} generate -g go-server --package-name v1 -i /local/${CLIENT_OPENAPIDEF_FILE} -o /local/tmpclient
	rm -r -f pkg/bridgeclient/generated/v1
	mkdir -p pkg/bridgeclient/generated/v1
	cp tmpclient/go/model* pkg/bridgeclient/generated/v1

incontainergenerate:
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
buildmac: basebuild


buildlinux:	export GOOS=linux
buildlinux: export GOARCH=amd64
buildlinux: export CGO_ENABLED=0
buildlinux: export GO111MODULE=on
buildlinux: basebuild

build: basebuild

basebuild:
	mkdir -p out
	rm -f  out/bridgeserver_x64_linux
	go build -ldflags "-X github.com/theotw/natssync/pkg.VERSION=${BUILD_VERSION}" -v -o out/bridgeserver_${GOARCH}_${GOOS} apps/bridge_server.go
	go build -ldflags "-X github.com/theotw/natssync/pkg.VERSION=${BUILD_VERSION}" -v -o out/bridgeclient_${GOARCH}_${GOOS} apps/bridge_client.go
	go build -ldflags "-X github.com/theotw/natssync/pkg.VERSION=${BUILD_VERSION}" -v -o out/echo_main_${GOARCH}_${GOOS} apps/echo_main.go
	go build -ldflags "-X github.com/theotw/natssync/pkg.VERSION=${BUILD_VERSION}" -v -o out/echo_client_${GOARCH}_${GOOS} apps/echo_client.go
	go build -ldflags "-X github.com/theotw/natssync/pkg.VERSION=${BUILD_VERSION}" -v -o out/simple_auth_${GOARCH}_${GOOS} apps/simple_reg_auth_server.go
	go build -ldflags "-X github.com/theotw/natssync/pkg.VERSION=${BUILD_VERSION}" -v -o out/http_proxy_${GOARCH}_${GOOS} apps/httpproxy_server.go
	go build -ldflags "-X github.com/theotw/natssync/pkg.VERSION=${BUILD_VERSION}" -v -o out/http_proxylet_${GOARCH}_${GOOS} apps/http_proxylet.go

buildarm: export GOOS=linux
buildarm: export GOARCH=arm
buildarm: export CGO_ENABLED=0
buildarm: export GO111MODULE=on
buildarm: basebuild

clean:
	rm -r -f tmp
	rm -r -f pkg/bridgemodel/generated/v1
	rm -r -f out
	rm go.sum

baseimage:
	docker build  --no-cache --tag natssync-base:latest -f Dockerfilebase .
baseimagearm:
	docker build --tag natssync-base:arm-latest -f DockerfilebaseArm .

cloudimage:
	DOCKER_BUILDKIT=1 docker build --no-cache --build-arg IMAGE_REPO=${IMAGE_REPO} --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/natssync-server:${IMAGE_TAG} --target natssync-server .
cloudimageBuildAndPush: cloudimage
	docker push ${IMAGE_REPO}/natssync-server:${IMAGE_TAG}
cloudimagearm:
	DOCKER_BUILDKIT=1 docker build --build-arg IMAGE_REPO=${IMAGE_REPO} --build-arg IMAGE_TAG=${IMAGE_TAG} -f DockerfileArm --tag ${IMAGE_REPO}/natssync-server:arm-${IMAGE_TAG} --target natssync-server-arm .

debugcloudimage:
	DOCKER_BUILDKIT=1 docker build --build-arg IMAGE_REPO=${IMAGE_REPO} --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/natssync-server-debug:${IMAGE_TAG} --target debugnatssync-server .


testimage:
	docker build --build-arg IMAGE_REPO=${IMAGE_REPO} --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/natssync-tests:${IMAGE_TAG} --target natssync-tests .
testimageBuildAndPush: testimage
	docker push ${IMAGE_REPO}/natssync-tests:${IMAGE_TAG}

clientimage:
	DOCKER_BUILDKIT=1 docker build --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/natssync-client:${IMAGE_TAG} --target natssync-client .
clientimageBuildAndPush: clientimage
	docker push ${IMAGE_REPO}/natssync-client:${IMAGE_TAG}
clientimagearm:
	DOCKER_BUILDKIT=1 docker build -f DockerfileArm --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/natssync-client:arm-${IMAGE_TAG} --target natssync-client-arm .

echoproxylet:
	DOCKER_BUILDKIT=1 docker build --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/echo-proxylet:${IMAGE_TAG} --target echo-proxylet .
echoproxyletBuildAndPush: echoproxylet
	docker push ${IMAGE_REPO}/echo-proxylet:${IMAGE_TAG}

echoproxyletarm:
	DOCKER_BUILDKIT=1 docker build -f DockerfileArm --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/echo-proxylet:arm-${IMAGE_TAG} --target echo-proxylet-arm .

simpleauth:
	DOCKER_BUILDKIT=1 docker build --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/simple-reg-auth:${IMAGE_TAG} --target simple-reg-auth .
simpleautharm:
	DOCKER_BUILDKIT=1 docker build -f DockerfileArm --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/simple-reg-auth:arm-${IMAGE_TAG} --target simple-reg-auth-arm .

simpleauthBuildAndPush: simpleauth
	docker push ${IMAGE_REPO}/simple-reg-auth:${IMAGE_TAG}

httpproxy:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/httpproxy-server:${IMAGE_TAG} --target http_proxy .

httpproxylet:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile --build-arg IMAGE_TAG=${IMAGE_TAG} --tag ${IMAGE_REPO}/httpproxylet:${IMAGE_TAG} --target http_proxylet .

nginxTest:
	cd testNginx && docker build --tag ${IMAGE_REPO}/testnginx:${IMAGE_TAG} .

allimages: baseimage testimage cloudimage clientimage echoproxylet simpleauth debugcloudimage httpproxy httpproxylet

allarmimages: baseimagearm cloudimagearm clientimagearm echoproxyletarm simpleautharm

allimagesBuildAndPush:testimageBuildAndPush cloudimageBuildAndPush clientimageBuildAndPush echoproxyletBuildAndPush simpleauthBuildAndPush

imagelist=natssync-server natssync-client echo-proxylet simple-reg-auth natssync-tests natssync-server-debug httpproxy-server httpproxylet
loopover:
	@ for img in ${imagelist}; \
 		do \
 			echo $${img}; \
 		done
buildAndTag: allimages tag
tag:
	docker tag ${IMAGE_REPO}/natssync-server:${IMAGE_TAG} ${IMAGE_REPO}/natssync-server:latest
	docker tag ${IMAGE_REPO}/natssync-client:${IMAGE_TAG} ${IMAGE_REPO}/natssync-client:latest
	docker tag ${IMAGE_REPO}/echo-proxylet:${IMAGE_TAG} ${IMAGE_REPO}/echo-proxylet:latest
	docker tag ${IMAGE_REPO}/simple-reg-auth:${IMAGE_TAG} ${IMAGE_REPO}/simple-reg-auth:latest
	docker tag ${IMAGE_REPO}/natssync-tests:${IMAGE_TAG} ${IMAGE_REPO}/natssync-tests:latest
	docker tag ${IMAGE_REPO}/natssync-server-debug:${IMAGE_TAG} ${IMAGE_REPO}/natssync-server-debug:latest
	docker tag ${IMAGE_REPO}/httpproxy-server:${IMAGE_TAG} ${IMAGE_REPO}/httpproxy-server:latest
#   httpproxy_server is depricated use httpproxy-server instead 
	docker tag ${IMAGE_REPO}/httpproxy-server:${IMAGE_TAG} ${IMAGE_REPO}/httpproxy_server:latest
	docker tag ${IMAGE_REPO}/httpproxylet:${IMAGE_TAG} ${IMAGE_REPO}/httpproxylet:latest

tagAndPushToDockerHub: tag
	docker push ${IMAGE_REPO}/natssync-server:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/natssync-server:latest
	docker push ${IMAGE_REPO}/natssync-client:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/natssync-client:latest
	docker push ${IMAGE_REPO}/echo-proxylet:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/echo-proxylet:latest
	docker push ${IMAGE_REPO}/simple-reg-auth:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/simple-reg-auth:latest
	docker push ${IMAGE_REPO}/natssync-tests:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/natssync-tests:latest
	docker push ${IMAGE_REPO}/natssync-server-debug:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/natssync-server-debug:latest
	docker push ${IMAGE_REPO}/httpproxy-server:${IMAGE_TAG}
 	docker push ${IMAGE_REPO}/httpproxy-server:latest
#   httpproxy_server is depricated use httpproxy-server instead 
	docker push ${IMAGE_REPO}/httpproxy_server:${IMAGE_TAG}
 	docker push ${IMAGE_REPO}/httpproxy_server:latest
	docker push ${IMAGE_REPO}/httpproxylet:${IMAGE_TAG}
 	docker push ${IMAGE_REPO}/httpproxylet:latest



pushall:
	docker push ${IMAGE_REPO}/natssync-server:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/natssync-client:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/echo-proxylet:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/simple-reg-auth:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/natssync-tests:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/httpproxy-server:${IMAGE_TAG}
#   httpproxy_server is depricated use httpproxy-server instead 
	docker push ${IMAGE_REPO}/httpproxy_server:${IMAGE_TAG}
	docker push ${IMAGE_REPO}/httpproxylet:${IMAGE_TAG}


l1:
	SUCCESS=0; \
	go get github.com/jstemmer/go-junit-report; \
	mkdir -p out; \
	go test -v -coverpkg=github.com/theotw/natssync/pkg/... -coverprofile=out/unit_coverage.out github.com/theotw/natssync/pkg/... > out/l1_out.txt 2>&1 || SUCCESS=1; \
	cat out/l1_out.txt | go-junit-report > out/l1_report.xml || echo "Failure generating report xml"; \
	cat out/l1_out.txt; \
	exit $$SUCCESS;

integrationtests: SYNCCLIENT_PORT ?= 8081
integrationtests:
	./single_cluster_test/docker-cleanup.sh && SYNCCLIENT_PORT=${SYNCCLIENT_PORT} ./single_cluster_test/docker-deploy.sh "${IMAGE_TAG}"
	go run apps/natstool.go -u nats://localhost:4222 -s natssync.registration.request -m '{"authToken":"42","locationID":"client1"}'
	sleep 10
	curl -X POST -H 'Content-Type: application/json' -d '{"authToken":"42","locationID":"client1"}' "http://localhost:${SYNCCLIENT_PORT}/bridge-client/1/register" | jq .locationID | sed s/\"//g > locationID.txt
	echo "ID: `cat locationID.txt`"
	go run apps/echo_client.go -m "hello world" -i `cat locationID.txt`
	syncserver_url='http://localhost:8080' syncclient_url='http://localhost:${SYNCCLIENT_PORT}' natsserver_url='nats://localhost:4222' go test -v github.com/theotw/natssync/tests/integration/...
	curl -i -f -X POST -H 'Content-Type: application/json' -d '{"authToken":"42","locationID":"`cat locationID.txt`"}' "http://localhost:8081/bridge-client/1/unregister"
	echo "Unregistered ID: `cat locationID.txt`"
	echo "Single cluster test done"

coveragereport:
	./scripts/exit_apps_gracefully.sh
	gocovmerge out/*_coverage.out > out/merged.out
	go tool cover -func out/merged.out

writeimage:
	$(shell echo ${IMAGE_TAG} >'IMAGE_TAG')
cicd: allimages

