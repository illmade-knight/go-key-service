# **go-key-service**

The Key Service is a secure, durable, and highly available microservice for storing and retrieving entity public keys. It serves as a foundational component in a secure messaging ecosystem, enabling clients to discover keys for end-to-end encryption.

The service is built on the standardized go-microservice-base library, ensuring consistent lifecycle management, observability, and error handling across our entire fleet of microservices. It features a clean, layered architecture that separates its public API from internal implementation details, allowing for flexible data storage backends and robust testing.

## **Directory Structure**

The repository follows standard Go project layout conventions for clarity and maintainability.

.  
├── cmd/  
│   └── keyservice/  
│       └── main.go              \# Assembles dependencies and runs the service  
├── internal/  
│   ├── api/                     \# Private HTTP handlers and middleware  
│   │   └── ...  
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

## **Current State: Production Ready**

The service is feature-complete according to its development roadmap and is ready for production deployment. It has been fully refactored to incorporate our standard microservice base library and the latest security model.

### **Features Implemented**

* ✅ **Standardized Service Base**: The service now uses go-microservice-base, providing robust, consistent lifecycle management and a graceful shutdown sequence.
* ✅ **Full Observability**: Exposes standard endpoints for monitoring:
    * GET /healthz: Liveness probe to confirm the service is running.
    * GET /readyz: Readiness probe to confirm the service is ready to handle traffic.
    * GET /metrics: Exposes performance metrics in the Prometheus format.
* ✅ **Secure JWT Authentication (RS256)**: The POST /keys/{entityURN} endpoint is secured. The service validates asymmetric RS256 tokens by fetching public keys from the identity service's JWKS endpoint.
* ✅ **Robust Authorization**: A user can only store a key for themselves, enforced by matching the JWT sub claim against the entity ID in the URN.
* ✅ **URN-Based Identity**: The service can store and retrieve keys for any entity type (users, devices, etc.) using a generic Uniform Resource Name (URN) identifier.
* ✅ **Persistent Storage**: A production-ready FirestoreStore provides a durable backend for storing keys. An InMemoryStore is available for testing.
* ✅ **Structured Error Handling**: All API errors are returned as standardized {"error": "message"} JSON objects.
* ✅ **Structured Logging**: All logging is handled by zerolog for machine-readable output.

## **Deployment and Running**

This service is designed to run on Google Cloud.

1. **Set Environment Variables**:  
   export GCP\_PROJECT\_ID=my-gcp-project  
   export IDENTITY\_SERVICE\_URL="http://localhost:3000"

   In production, these secrets are injected from Google Secret Manager but are consumed by the application as standard environment variables.
2. **Authenticate with Google Cloud**:  
   gcloud auth application-default login

3. **Enable Firestore API** in your GCP project and create a database.
4. **Run the Service**:  
   go run cmd/keyservice/runscalablekeyservice.go

5. **API Usage**:  
   \# Store a key (requires a valid JWT for the user)  
   $ curl \-X POST \-H "Authorization: Bearer \<JWT\>" \-d "THIS\_IS\_A\_KEY" http://localhost:8081/keys/urn:sm:user:user-alice

   \# Retrieve a key (publicly accessible)  
   $ curl http://localhost:8081/keys/urn:sm:user:user-alice  
   THIS\_IS\_A\_KEY  
