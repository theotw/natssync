# Single Cluster Testing Environment
These files and scripts are meant to help deploy a
testing environment in a single kubernetes cluster.

## Deployment
If you're using minikube, set your docker env for the minikube instance
```shell
eval $(minikube -p minikube docker-env)
```

From the root of the repo, run the `make` command to build all the images
```shell
make allimages
```

From the `single_cluster_test` directory, run the `deploy.sh` script
```shell
./deploy.sh
```

## Namespaces
The deploy script will set up two different namespaces.
- `cloud`
  The environment running in the cloud
- `onprem`
  The environment running on prem

## Ports
Several NodePort services have been deployed that expose ports

| Service | Description | Port |
| ------- | ----------- | ---- |
| nats | NATS in the cloud | 32222 |
| sync-server | NATSSync Server | 30080 |
| sync-client | NATSSync Client | 30081 |

**NOTE:** If the cluster is deployed in minikube,
you will need to run minikube commands to expose these services.
ex:
```shell
minikube service nats -n cloud
```

## Registration
At this point, everything is up and running.
However, the client side must be registered with the cloud side before they
can communicate.
Use the `natstool.go` tool under apps to send a registration request.
```shell
go run natstool.go -u nats://localhost:32222 -s natssync.registration.request -m '{"authToken":"42","locationID":"client1"}' -r response-subject
```

On the client side, you will need provide the same credentials.
```shell
curl -X POST -H 'Content-Type: application/json' -d '{"authToken":"42","locationID":"client1"}' http://localhost:30081/bridge-client/1/register
```
or with the SwaggerUI:
```shell
# URL
http://localhost:30081/bridge-client/api/index_bridge_client_v1.html#/default/post_register

# POST JSON
{
  "metaData": "client1",
  "authToken": "42"
}
```

## Connection Testing
You should now be able to use the `echo_client.go` tool to test the connection.
```shell
# The ID (-i option) will be unique to your instance.
# Check the logs of your sync-server to find your ID.
go run echo_client.go -u nats://localhost:32222 -i aad404f9-44cc-469b-82c9-d5c1d82275a1 -m "hello world"
```

If successful, your output should look something like this:
```
2021/02/10 16:27:45 Connecting to NATS Server nats://localhost:32222
2021/02/10 16:27:45 Subscribed to natssync-nb.cloud-master.68181010-0f03-4d0e-b7f7-f7a2381bb5e2.*
2021/02/10 16:27:45 Published message: hello world
2021/02/10 16:27:45 Message received [natssync-nb.cloud-master.68181010-0f03-4d0e-b7f7-f7a2381bb5e2.bridge-msg-handler]: 20210210-23:27:45.428 | message-handler

2021/02/10 16:27:46 Message received [natssync-nb.cloud-master.68181010-0f03-4d0e-b7f7-f7a2381bb5e2.bridge-client]: 20210210-23:27:46.517 | message-client

2021/02/10 16:27:46 Message received [natssync-nb.cloud-master.68181010-0f03-4d0e-b7f7-f7a2381bb5e2.echolet]: 20210210-23:27:46.519 | echoproxylet * hello world
```

## Cleanup
To take down the environment, run the cleanup script.
This will simply delete the cloud and onprem namespaces.
```shell
./clean.sh
```
