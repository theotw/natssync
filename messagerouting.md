#Message Routing
## Overview
This documents details how message subjects need to be formatted to be delivered through NATSSync 

As a general rule, 2 or more services listening for messages should work if all services share the same NATS or if they share different NATS.

There are a coupe conventions for message format and selection that must be followed to allow this to happen

## Message to be sent via sync
Messages to be sent through the NATSSync system need to be prefix with:
`natssync.<locationID>.`
* Where  location ID is the final destination of the message.

The location ID of 1 is reserved for the natssync server.

## Routing
Message prefixed with a valid natssync will be sent to the client for the given location ID.

### NATSSYNC Server
The server holds queue for all known clients.  
When a client request messages for itself, the server will pull from the prooper queue

When the server is sent a message via its POST interfaces, then it will place
those messages with valid subjects onto the NATS System

### NATSSync Client
THe NATSSync client listens for ALL messages prefix with the natssync prefix.
If the message has a local location ID, that message is processed as any other message targeted for the client.

If the message has an different location ID, it is passed to the server.

Messages the client receives from the server are are posted to the local NATS System

## System Flag subject segment
This segment of the subject should be set to match everything *

Services sending data MUST set this field to 1

However, service should not select on this field, the system WILL change it.  Service should wild card it


Service receiving messages should wild card and ignore it

All services should assume that value will be modified by the NATSSYnc System

The NATSSync Client will slip the system flag from 1 to 0 when it puts a message it pulls from the server to the local NATS

This will stop it from picking up and sending it to the server.
