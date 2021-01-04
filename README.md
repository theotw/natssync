# natssync
A distributed bridge system to sync messages from one nats cluster to another.  The idea is to use this across the Cloud / On Prem gap
With message encryption


## Env Vars
| Env Var | Default | Description |
| ---- | ------------ | ----------- |
| REDIS_URL | localhost:6379 | url to redis |
| NATS_SERVER_URL | localhost:4222 | the nats server to use
| CACHE_MGR | redis | name of the cache mgr, redis and mem are currently supported
| KEYSTORE | redis | name of the keystore to use, file or redis 
 
