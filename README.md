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
| LISTEN_STRING | :8080 | the port on which the rest api listens
 

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

# Tagging and Versioning

Version is controlled by the version.txt for the major.minor and a build number which is a date string

Each trunk/main build, images are pushed to docker hub.

Images pushed are timestamped for the build number and then the latest label is placed in the most recent. A major.minor
version label is placed on the most recent build for that version. Users can tag to the major / minor version to get bug
fix builds
