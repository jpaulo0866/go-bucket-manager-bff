# 0004. Strategy Pattern for Cloud Storage Services

- **Date**: 2025-06-20
- **Status**: Accepted

## Context

The application must interact with multiple cloud storage providers (AWS, GCP, Azure). Each provider has its own distinct SDK and API for performing operations like listing files, downloading, uploading, and generating presigned URLs.

A naive implementation within a single service class would require complex conditional logic (e.g., `if/else` or `switch` statements) to select the correct provider's code path. This approach would violate the Open/Closed Principle, making the service difficult to read, maintain, and extend with new providers in the future. We need a design that is clean, scalable, and decouples the core business logic from the specific implementation details of each cloud provider.

## Decision

We will use the **Strategy design pattern** to manage interactions with the different cloud storage providers.

1. **Strategy Interface**: We define a common `StorageStrategy` interface. This interface declares the contract for all storage operations (e.g., `ListFiles`, `DownloadFile`, `UploadFile` and `PresignedURL`).

2. **Concrete Strategies**: For each supported cloud provider, we create a separate class that implements the `StorageStrategy` interface (e.g., `GCSStrategy`, `S3Strategy`, `AzureStrategy`). Each class contains the provider-specific logic and uses the corresponding SDK.

3. **Conditional Creation**: Each strategy implementation is only instantiated if the corresponding cloud provider is enabled in the configuration (e.g., `AWS_ENABLED`, `GCP_ENABLED`, `AZURE_ENABLED`).

4. **Strategy Context/Selector**: The `NewStrategies` function acts as the context. It is injected with all available `StorageStrategy` instances. It populates a `map[CloudProvider]StorageStrategy`, which allows for efficient look-up of the correct strategy at runtime based on the `CloudProvider` from the request. When a method called `NewStrategies`, it retrieves the appropriate strategies map and delegates the call.

## Consequences

### Positive

- **Open/Closed Principle**: The system is open for extension but closed for modification. To add support for a new cloud provider, we only need to create a new strategy class and add the validation to enable or not.
- **Decoupling**: The routing functions are completely decoupled from the low-level implementation details of any specific cloud SDK, promoting modularity and flexibility.
- **Improved Readability and Maintainability**: Logic for each provider is encapsulated in its own class, making the code easier to understand, test, and debug.
- **Testability**: Each strategy can be unit-tested in isolation.
- **Efficient Resource Management**: Using the validations with environment variables to instantiate the classes prevents unnecessary cloud SDKs and clients from being initialized, reducing the application's memory footprint and startup time.

### Negative

- **Increased Number of Classes**: The pattern results in a higher number of classes and interfaces in the project.
- **Indirection**: It introduces a layer of indirection. To understand the full execution flow, a developer must trace the call from the `NewStrategies` function to the strategy map and then to the concrete strategy implementation.
