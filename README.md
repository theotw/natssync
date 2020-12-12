# natssync
A distributed bridge system to sync messages from one nats cluster to another.  The idea is to use this across the Cloud / On Prem gap

## Env Vars
| Env Var | Default | Description |
| ---- | ------------ | ----------- |
| REDIS_URL | localhost:6379 | url to redis |
| CACHE_MGR | redis | name of the cache mgr, redis and mem are currently supported
| NATS_SERVER_URL | localhost:4222 | the nats server to use
 
