# **go-key-service**

The Key Service is a secure, durable, and highly available microservice for storing and retrieving users' public keys. It serves as a foundational component in a secure messaging ecosystem, enabling clients to discover keys for end-to-end encryption.

The service is built with a clean, layered architecture that separates its public API from internal implementation details. This decoupled design, consistent with our other microservices, allows for flexible data storage backends and robust testing.

---

## **Directory Structure**

The repository follows standard Go project layout conventions for clarity and maintainability.

Plaintext

.  
├── cmd/  
│   └── keyservice/  
│       └── main.go              \# Assembles dependencies and runs the service  
├── e2e/  
│   ├── full\_flow\_test.go        \# E2E test with routing-service (from other repo)  
│   └── fullkeystore\_test.go     \# Standalone E2E test for this service  
├── internal/  
│   ├── api/  
│   │   └── handlers.go          \# Private HTTP handlers  
│   └── storage/  
│       ├── firestore/           \# Concrete Firestore store implementation  
│       └── inmemory/            \# Concrete in-memory store implementation  
├── pkg/  
│   └── keyservice/  
│       ├── config.go            \# Public Config struct  
│       └── store.go             \# Public Store interface  
├── keyservice/  
│   └── service.go               \# The public, embeddable service library/wrapper  
└── test/  
└── e2e\_helpers.go           \# Public helpers for E2E tests

* **cmd/**: Contains the executable entry point. Its main.go file assembles the service by creating concrete dependencies (like a Firestore client) and injecting them into the service wrapper.
* **e2e/**: Holds end-to-end integration tests that treat the service as a black box, verifying its functionality via its public HTTP API.
* **internal/**: Contains all private application logic. This includes the HTTP handlers (api) and the concrete storage implementations (storage/firestore, storage/inmemory).
* **pkg/**: Holds the service's public contract. pkg/keyservice defines the Store interface and Config struct that other packages use to interact with the service's domain.
* **keyservice/**: The primary, public-facing service library. It provides the Wrapper and New() constructor, which assembles the internal components into a runnable service.
* **test/**: Provides public helper functions that allow external E2E tests to create a fully assembled test server without violating the internal package boundary.

---

## **Current State vs. Roadmap**

The service has been successfully refactored, completing the foundational work of **Phase 1** of the development roadmap. The architecture is now robust, testable, and ready for production-hardening features.

## **Deployment and running**

This service is implemented to run on Google Cloud, this assumes you have gcloud cli installed, have generated credentials (using `gcloud auth login`). The command to start the service is also in a Makefile

* Create a Google Cloud Project and set GCP_PROJECT_ID
```
export GCP_PROJECT_ID=myproject-mvp
```
* Get credentials to be used in calling Google APIs
```
gcloud auth application-default login
```
* In order to use GCloud Firestore you will need to enable the Cloud Firestore API by visiting this api
https://console.cloud.google.com/apis/api/firestore.googleapis.com/metrics?project=myproject-mvp
	* You will need to create a default database as well by visiting this url - you'll need to select some options such as region, encryption, SLA etc
	https://console.cloud.google.com/datastore/setup?project=homesafemvp
* Run the key service using GCloud Firestore as storage
```
go run cmd/keyservice/runscalablekeyservice.go
```
* Run the key service using in memory storage
```
go run cmd/simple/runkeystore.go
```
* Key API now avail at http://localhost:8081/keys/ e.g.
```
$ curl -d "THIS_IS_A_KEY" http://localhost:8081/keys/urn:sm:user:user-alice
$ curl  http://localhost:8081/keys/urn:sm:user:user-alice
THIS_IS_A_KEY
```


### **What's Complete (Phase 1 Foundation)**

The core architecture and primary features are now implemented and tested.

* ✅ **Architectural Refactor:** The project has been fully migrated to the clean, layered architecture.
* ✅ **Persistent Storage:** The FirestoreStore has been implemented, providing a durable, production-ready backend for storing keys.
* ✅ **API Layer:** The POST /keys/{userID} and GET /keys/{userID} endpoints are fully functional.
* ✅ **Dependency Injection:** The main.go executable correctly assembles the service and injects the chosen storage backend.
* ✅ **Configuration:** The service's listen address is now managed via the Config struct.
* ✅ **Logging:** All logging is now handled by zerolog for structured, machine-readable output.

### **What's Next (Phase 2 & 3\)**

The immediate next steps involve implementing the production-hardening and security features outlined in the roadmap.

* **Authentication & Authorization:** The highest priority is to implement **JWT-based authentication and authorization** to secure the API endpoints. A user should only be able to manage their own key.
* **Observability:** While structured logging is in place, the service still needs a /healthz endpoint and Prometheus metrics for monitoring.
* **Improved Error Handling:** API errors should be returned as structured JSON, not plain text.
* **Deployment:** The service needs a Dockerfile for containerization and a CI/CD pipeline for automated builds and deployments.