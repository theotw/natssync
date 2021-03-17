I'm using K3D to simplify simulations of the metrics.  If you don't have K3D on a Mac do:
```bash
brew install k3d
```
Start a simple K3D cluster to suit our needs using:
```bash
./tests/local/k3d.sh
```
The startup is script does not do any error handling.  
It either creates a new k3d cluster or deletes the existing cluster and creates a new one.
It should automatically change your context to `k3d-nats`.  If it doesn't:
```bash
kubectl config use-context k3d-nats
```
The script should start a local docker registry where you can store your built containers. 
After k3d starts find it by using this command:
```bash
docker ps -f name=k3d-nats-registry
```

If you encounter the error `Warning   InvalidDiskCapacity       node/k3d-nats-agent-0    invalid capacity 0 on image filesystem` 
This can be fixed by flushing the docker image cache `docker system prune -a` followed by re-running `k3d.sh`

When k3d is running, deploy a prometheus stack using:
```bash
kubectl apply -f prometheus.yaml
```
I did not configure this with any persistent storage which means you should **connect the prometheus data source at first run in grafana**.
Grafana is configured with no authentication.
After the data source exists you can import the defined graph json files such as `NATSsyncDevelopmentDashboard.json`.
*The alertmanger config is currently empty and therefore broken but this is ok for now*

Add a line for the local registry to /etc/hosts
```bash
127.0.0.1 k3d-nats-registry
```

If local firewall nonsense is getting you down, turn off the firewall like so:
```bash
sudo defaults write /Library/Preferences/com.apple.alf globalstate -int 0
```

Build and tag and push locally your images(obviously you may need to pick a different port from the commands above):
```bash
make cloudimage clientimage
docker tag theotw/natssync-client:latest k3d-nats-registry:55519/theotw/natssync-client:latest
docker push k3d-nats-registry:55519/theotw/natssync-client:latest
docker tag theotw/natssync-client:latest k3d-nats-registry:55519/theotw/natssync-server:latest
docker push k3d-nats-registry:55519/theotw/natssync-server:latest
```

You should be able to use the simple deployment within your k3d cluster now:
```bash
kubectl apply -f deployment.yaml
```

Based on the port forwards in k3d.yaml you should be able to see metrics here:
```bash
curl http://localhost:9082/metrics
curl http://localhost:9081/onprem-bridge/metrics
```

What remains to be done is to find an efficient way of hitting the pods with various messages and watching how it fails.