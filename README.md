## Couchbase storage for osin OAuth server library

Provides storage for [osin](https://github.com/RangelReale/osin)

### Installation

```
go get github.com/elgris/osin-couchbase-storage
```

### Running tests

Tests are not pure unit tests, they require running Couchbase instance. You can provide all necessary connection parameters via command line:
```
go test -couchbase="couchbase://localhost" -bucket="test" -password="111"
```

### Using with osin

Example:

```
import (
    "github.com/elgris/osin-couchbase-storage"
    "github.com/RangelReale/osin"
)

s, err := storage.NewStorage(storage.Config{
    ConnectionString: "couchbase://localhost",
    BucketName:       "default",
    BucketPassword:   "",
})

if err != nil {
    panic(err.Error())
}

server := osin.NewServer(osin.NewServerConfig(), s)
```

### License

MIT
