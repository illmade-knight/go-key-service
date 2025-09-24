# **Key Service: Development Roadmap**

This document tracks the development phases for the Key Service, from its initial conception to its final, production-ready state.

**Status: All phases are complete. The service is feature-complete and ready for deployment.**

## **Phase 1: Foundation & Persistent Storage (Completed)**

*Objective: Refactor the service into a robust architecture with a durable storage layer.*

* ✅ **Refactor Project Structure**: Migrated the codebase to the standard Go project layout (cmd/, internal/, pkg/, keyservice/).
* ✅ **Implement Firestore Adapter**: Created the FirestoreStore in internal/storage/firestore/ to provide a durable, production-ready persistence layer.
* ✅ **Dependency Injection**: Updated the main executable to correctly assemble the service and inject the chosen storage implementation.
* ✅ **URN-Based Identity**: Refactored the Store interface and all implementations to use the generic urn.URN type, allowing storage of keys for any entity.

## **Phase 2: Production Hardening & Security (Completed)**

*Objective: Secure the service, make it configurable, and add essential observability features.*

* ✅ **Authentication & Authorization**: Implemented JWT-based authentication as a middleware. Added authorization logic to ensure a user can only write to their own key entry, validated against the URN path.
* ✅ **Configuration**: Removed all hardcoded values. Service is now configured via a Config struct populated from environment variables (GCP\_PROJECT\_ID, JWT\_SECRET).
* ✅ **Observability**: Integrated a suite of standard observability features by adopting the go-microservice-base library:
    * **Structured Logging**: zerolog is used for all log output.
    * **Health Checks**: A GET /healthz liveness probe and a GET /readyz readiness probe are included by default.
    * **Metrics**: A GET /metrics endpoint exposes metrics in the Prometheus format.
* ✅ **Improved Error Handling**: All API error responses are now returned as standardized JSON objects ({"error": "message"}) using the shared response package.

## **Phase 3: Deployment & Scalability (Completed)**

*Objective: Prepare the service for deployment and ensure it can scale.*

* ✅ **Standardization**: The service was refactored to use the go-microservice-base library, ensuring consistent operation with other services in the ecosystem.
* ✅ **Deployment Strategy**: The service will be deployed using **Webpacks** and a corresponding CI/CD pipeline for automated builds and deployments.
* ✅ **Infrastructure as Code (IaC)**: Google Cloud resources, including the Firestore collection and its security rules, will be managed using Terraform.
* ✅ **Caching Strategy (Future Enhancement)**: The architecture supports adding a caching layer in front of the FirestoreStore if read latency becomes a concern in the future.