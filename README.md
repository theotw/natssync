# NATSSync
A distributed bridge system to sync messages from one nats cluster to another.  The idea is to use this across the Cloud / On Prem gap
With message encryption


## Environment Variables
| Env Var | Default | Description |
| ---- | ------------ | ----------- |
| REDIS_URL | localhost:6379 | url to redis |
| NATS_SERVER_URL | localhost:4222 | the nats server to use
| CACHE_MGR | redis | name of the cache mgr, redis and mem are currently supported
| KEYSTORE | redis | name of the keystore to use, file or redis 
| GIN_PORT | 8080 | the port on which the rest api listens
 

## Testing
Unit tests can be run using the `l1` command from the `Makefile`
```shell
make l1
```


## Building the images
The cloud and client Docker image build commands are also in the `Makefile`
```shell
# NATS Sync Server Image
make cloudimage
```
```shell
# NATS Sync Client Image
make clientimage
```
