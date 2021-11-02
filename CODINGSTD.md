## Coding Standards

### HTTP Methods, status codes 
use constants defined in this library: 
https://pkg.go.dev/net/http#pkg-constants
instead of defining them as manually. 

### Log messages 
##### log messages with error:  

use: 
```
log.WithError(err).Errorf("Error encoding envelope to json bits")
```
instead of: 
```
log.Errorf("Error encoding envelope to json bits %s", jsonerr.Error())
```

##### log messages where particular field(s) are being logged: 

use: 
```
log.WithError(err).WithField("statusCode", statusCode).Errorf("Failed to rotate certificates")

log.WithFields(log.Fields{"animal": "walrus","size": 10}).Info("A group of walrus emerges from the ocean")
``` 

instead of: 
```
log.WithError(err).Errorf("Failed to rotate certificates: statusCode: %v", statusCode)

log.Info("A group of walrus emerges from the ocean: animal: %v size: %v",animalType, animalGroupSize )
```
