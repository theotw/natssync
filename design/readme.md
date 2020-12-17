# General Design Notes

# Message Convention
As a general rule there are no requirements on messages.  They are passed as is.  They are encrypted, but when they 
arrive at the NATS Message Server they are normal NATS messages.

1 Small item, Subject Names

Subject names are used for routing.  So they must have some basic form.  Messages to the client side (aka south side) must be prepended with natssync-sb

Messages coming back to the server (aka cloud side or north side) must be prepended with natssync-nb
 Then the location ID (on premise ID or the cloud server ID) must follow then the application specific data
 
 General for is:
 
 natssync-(sb|nb).locationID.application-specific-data
 
