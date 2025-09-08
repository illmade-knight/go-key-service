# **Key Service: Development Roadmap**

## **1\. Vision & Core Principles**

The **Key Service** is a secure, durable, and highly available microservice responsible for storing and retrieving users' public keys. It is a foundational component of the secure messaging ecosystem, enabling clients to discover each other's keys for encrypting messages.

* **Security First:** The service's primary responsibility is to protect the integrity of stored public keys. All endpoints must be authenticated and authorized.
* **Durability & Availability:** User keys are critical data and must never be lost. The service must be built on a persistent, replicated datastore and be highly available.
* **Simplicity:** The API and architecture should remain simple and focused on its core function: managing public keys.

---

## **2\. Core Architecture**

The service will be refactored into the same layered, decoupled architecture as the routing service. This will separate the public API, internal logic, and data storage layers, allowing for flexible and testable development.

* **Project Structure:** Adopt the standard Go project layout (cmd, internal, pkg, keyservice) to create clear boundaries.
* **Data Storage Strategy:** The current InMemoryStore is suitable only for local testing. The production implementation will use a durable, managed database.
    * **Firestore (Primary Persistent Store):** Firestore is an ideal choice for this use case. It's a managed, scalable, and durable NoSQL database that fits the key-document model (userID \-\> key) perfectly.
        * *Collection:* public\_keys
        * *Document ID:* \<userID\>
        * *Document Data:* A struct containing the public key bytes.
    * **In-Memory (Testing & Local Dev):** The existing InMemoryStore will be preserved as an internal implementation for use in unit tests and local development environments.

---

## **3\. Development Phases**

### **Phase 1: Foundation & Persistent Storage**

*Objective: Refactor the service into a robust architecture with a durable storage layer.*

1. **Refactor Project Structure:**
    * Create the new directory layout: cmd/, internal/, pkg/, keyservice/.
    * **pkg/keyservice/**: Define the public contracts. Move the storage.Store interface here.
    * **internal/storage/**: House the data layer implementations.
        * Move the existing InMemoryStore to internal/storage/inmemory/.
    * **internal/api/**: Create this package for the private HTTP handlers (the keyHandler).
    * **keyservice/**: Create the public service wrapper that assembles the components.
    * **cmd/keyservice/**: Create the main executable that performs dependency injection.
2. **Implement Firestore Adapter:**
    * Create a new FirestoreStore struct in internal/storage/firestore/ that implements the Store interface.
    * This adapter will handle all communication with the Firestore public\_keys collection.
3. **Dependency Injection:**
    * Update cmd/keyservice/main.go to choose the storage implementation based on configuration (e.g., use FirestoreStore in production, InMemoryStore if a flag is set).
    * The chosen Store instance will be injected into the API layer.

### **Phase 2: Production Hardening & Security**

*Objective: Secure the service, make it configurable, and add essential observability features.*

1. **Authentication & Authorization:**
    * Implement **JWT-based authentication** as a middleware for all /keys/ endpoints. Unauthenticated requests should be rejected.
    * Implement **Authorization**: Ensure that the userID from the authenticated JWT claim matches the \<userID\> in the URL path. A user must only be able to read or write their own key.
2. **Configuration:**
    * Remove all hardcoded values (e.g., the :8081 port).
    * Create a Config struct that is populated from environment variables or command-line flags.
3. **Observability:**
    * Integrate a structured logger (e.g., zerolog) to replace the standard log package.
    * Add basic Prometheus metrics for request latency and status codes.
    * Create a /healthz endpoint for health checks.
4. **Improved Error Handling:**
    * Return structured JSON error responses (e.g., {"error": "message"}) instead of plain text from http.Error.

### **Phase 3: Deployment & Scalability**

*Objective: Prepare the service for deployment and ensure it can scale.*

1. **Containerization:**
    * Write a Dockerfile to build a minimal, production-ready container image for the service.
2. **CI/CD:**
    * Set up a continuous integration pipeline (e.g., GitHub Actions) that runs tests, performs static analysis, and builds the container image on every push to the main branch.
3. **Infrastructure as Code (IaC):**
    * Use a tool like Terraform to manage the Google Cloud resources, including the Firestore collection and its security rules.
4. **Caching Strategy (Optional Future Enhancement):**
    * If read latency becomes a concern, implement a caching layer (e.g., using the RedisCache from go-dataflow) in front of the FirestoreStore to improve performance for frequently accessed keys.