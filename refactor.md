# **Key Service: Refactor Plan**

The objective is to evolve the key service from a single-file demo into a robust, scalable, and testable application. This plan follows the exact architectural patterns established in the routing service.

## **Core Principles**

The refactor will adhere to a standard Go project layout to ensure a clean separation of concerns:

* **pkg/**: Public, shareable libraries. For this service, it will define the core "contract"—the data storage interface and configuration structs.
* **internal/**: All private application logic, including API handlers and concrete storage implementations.
* **keyservice/**: The primary, public-facing service library. It provides an embeddable component that assembles and runs the service.
* **cmd/**: The thin executable wrapper that performs final dependency injection and starts the service.
* **test/ & e2e/**: Support for external, black-box end-to-end testing without violating internal boundaries.

## **Final Directory Structure**

Plaintext

key-service/  
├── cmd/  
│   └── keyservice/  
│       └── main.go              \# Assembles dependencies and runs the service  
├── internal/  
│   ├── api/  
│   │   └── handlers.go          \# Private HTTP handlers  
│   └── storage/  
│       └── inmemory/  
│           └── store.go         \# Concrete in-memory store implementation  
├── pkg/  
│   └── keyservice/  
│       ├── config.go            \# Public Config struct  
│       └── store.go             \# Public Store interface  
├── keyservice/  
│   └── service.go               \# The public, embeddable service wrapper  
└── go.mod

---

## **Step-by-Step Refactoring Guide**

### **Step 1: Reorganize Project Structure**

Create the new directories and move the existing code into them.

1. Move the Store interface from keystore.go into a new file at **pkg/keyservice/store.go**.
2. Move the InMemoryStore implementation from inmemorykeystore.go into **internal/storage/inmemory/store.go**.
3. The logic from runkeystore.go will be split into three new files: internal/api/handlers.go, keyservice/service.go, and cmd/keyservice/main.go.

### **Step 2: Define the Public Contract (pkg/keyservice/)**

Create the public, shareable components for the service.

**pkg/keyservice/store.go**

Go

package keyservice

// Store defines the public interface for key persistence.  
type Store interface {  
StoreKey(userID string, key \[\]byte) error  
GetKey(userID string) (\[\]byte, error)  
}

**pkg/keyservice/config.go**

Go

package keyservice

// Config holds all necessary configuration for the key service.  
type Config struct {  
HTTPListenAddr string  
}

### **Step 3: Create the API Layer (internal/api/)**

Isolate the HTTP handler logic into a private package.

**internal/api/handlers.go**

Go

package api

import (  
"io"  
"log"  
"net/http"

	"github.com/illmade-knight/key-service/pkg/keyservice" // Use the public interface  
)

// API holds the dependencies for the HTTP handlers.  
type API struct {  
Store keyservice.Store  
}

// KeyHandler manages GET and POST requests for user keys.  
func (a \*API) KeyHandler(w http.ResponseWriter, r \*http.Request) {  
userID := r.URL.Path\[len("/keys/"):\]

	switch r.Method {  
	case http.MethodPost:  
		key, err := io.ReadAll(r.Body)  
		if err \!= nil {  
			http.Error(w, "Cannot read request body", http.StatusBadRequest)  
			return  
		}  
		if err := a.Store.StoreKey(userID, key); err \!= nil {  
			http.Error(w, "Failed to store key", http.StatusInternalServerError)  
			return  
		}  
		w.WriteHeader(http.StatusCreated)  
		log.Printf("Stored public key for user: %s", userID)

	case http.MethodGet:  
		key, err := a.Store.GetKey(userID)  
		if err \!= nil {  
			http.NotFound(w, r)  
			return  
		}  
		w.Header().Set("Content-Type", "application/octet-stream")  
		w.Write(key)  
		log.Printf("Served public key for user: %s", userID)  
	}  
}

### **Step 4: Create the Public Service Wrapper (keyservice/)**

This is the primary, importable library for running the service.

**keyservice/service.go**

Go

package keyservice

import (  
"context"  
"errors"  
"log"  
"net/http"

	"github.com/illmade-knight/key-service/internal/api"  
	"github.com/illmade-knight/key-service/pkg/keyservice"  
)

// Wrapper encapsulates all components of the running service.  
type Wrapper struct {  
cfg    \*keyservice.Config  
server \*http.Server  
}

// New creates and wires up the entire key service.  
func New(cfg \*keyservice.Config, store keyservice.Store) \*Wrapper {  
apiHandler := \&api.API{Store: store}

	mux := http.NewServeMux()  
	mux.HandleFunc("/keys/", apiHandler.KeyHandler)

	return \&Wrapper{  
		cfg:    cfg,  
		server: \&http.Server{Addr: cfg.HTTPListenAddr, Handler: mux},  
	}  
}

// Start runs the service's HTTP server.  
func (w \*Wrapper) Start() error {  
log.Printf("Key Service listening on %s...", w.cfg.HTTPListenAddr)  
if err := w.server.ListenAndServe(); \!errors.Is(err, http.ErrServerClosed) {  
return err  
}  
return nil  
}

// Shutdown gracefully stops the service.  
func (w \*Wrapper) Shutdown(ctx context.Context) error {  
return w.server.Shutdown(ctx)  
}

### **Step 5: Create the Main Executable (cmd/keyservice/)**

This thin wrapper assembles and runs the service.

**cmd/keyservice/main.go**

Go

package main

import (  
"log"

	inmemorystore "github.com/illmade-knight/key-service/internal/storage/inmemory"  
	"github.com/illmade-knight/key-service/keyservice"  
	ks "github.com/illmade-knight/key-service/pkg/keyservice"  
)

func main() {  
// 1\. Load configuration  
cfg := \&ks.Config{  
HTTPListenAddr: ":8081",  
}

	// 2\. Choose and create the concrete storage implementation  
	//    (Here we use in-memory, but this could be Firestore based on config)  
	store := inmemorystore.NewInMemoryStore()

	// 3\. Create the service wrapper, injecting the store  
	service := keyservice.New(cfg, store)

	// 4\. Start the service  
	if err := service.Start(); err \!= nil {  
		log.Fatalf("Failed to start key service: %v", err)  
	}  
}  
