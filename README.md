# tracing

This package provides tracing functionality middleware.  
It manages headers and context.  

### To install tracing:
```go
go get -u github.com/andriiyaremenko/tracing
```

### How to use:

```go
package main

import (
	"net/http"

	"github.com/andriiyaremenko/tracing"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func main() {
	r := chi.NewRouter()
	r.Use(tracing.Middleware(tracing.DefaultMetadataOptions, uuid.NewString))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	http.ListenAndServe(":3000", r)
}
```
