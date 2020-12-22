# natssync
A distributed bridge system to sync messages from one nats cluster to another.  The idea is to use this across the Cloud / On Prem gap
With message encryption

#Design
## Big Parts
Pretty standard piece of client/server software

We have the bridge client which runs on prem to grab south bound messages from the bridge server.

Cloud Server runs on the North Bound side (cloud side) and listens for messages on a NATS system that are meant to go to the client side (Based on subject)

The client connects to the servers, ask for any messages for the client, it downloads them and puts them onto the NATS queue locally.
The client also pushes any messages going north bound.

Main packages
 * bridgeclient - Classes for the client (south) side of the communication
 * bridgeserver - classes for the bridge (north) side of the communications
 * bridgemodel - the glue that binds client and bridge server
 * msgs - Classes for storing and encrypting messages
 
## Plugable Parts
 * The bridge server stores messages heading south for the client.  That message cache is pluggable.  We have an in mem (no good for HA) and a redis cache.  
 * The msgs packages has a key store interface.  We have an file based (no good for HA, but maybe fine for the client)  and a REDIS key store

## Env Vars
| Env Var | Default | Description |
| ---- | ------------ | ----------- |
| REDIS_URL | localhost:6379 | url to redis |
| CACHE_MGR | redis | name of the cache mgr, redis and mem are currently supported
| NATS_SERVER_URL | localhost:4222 | the nats server to use
 
