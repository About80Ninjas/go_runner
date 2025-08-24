# Go Runner

Go Runner is a powerful and flexible REST API service designed for managing and executing Go binaries built directly from Git repositories. This service provides a seamless workflow for dynamically building, deploying, and running Go programs through a comprehensive REST API, complete with admin and API documentation interfaces.

## 🌟 Features

- **Dynamic Go Builds**: Build Go applications from any Git repository and branch.
- **Remote Execution**: Execute pre-compiled binaries with custom arguments, environment variables, and stdin.
- **RESTful API**: A complete API for managing the lifecycle of binaries and their execution.
- **Secure**: Protect your endpoints with API keys for execution and an admin token for management.
- **Containerized**: Ready for deployment with Docker and Docker Compose.
- **Admin & Docs UI**: Comes with a built-in admin interface and Swagger/OpenAPI documentation.
- **Configuration**: Easily configurable through environment variables.
- **Scalable**: Designed to handle concurrent executions with configurable limits.

## 🚀 Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

### Running with Docker Compose

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/your-username/go_runner.git
    cd go_runner
    ```

2.  **Create a `.env` file:**

    Create a `.env` file in the root of the project and add the following environment variables. For production use, be sure to use a strong, randomly generated `ADMIN_TOKEN`.

    ```env
    ADMIN_TOKEN=your-secret-admin-token
    ```

3.  **Start the service:**

    ```bash
    docker-compose up -d
    ```

4.  **Access the service:**

    -   **API**: `http://localhost:8080/api/v1`
    -   **Swagger Docs**: `http://localhost:8080/api/v1/docs`
    -   **Admin UI**: `http://localhost:8080/admin`

## ⚙️ Configuration

The service is configured using environment variables. You can set these in your `docker-compose.yml` or a `.env` file.

| Variable                 | Description                                       | Default                  |
| ------------------------ | ------------------------------------------------- | ------------------------ |
| `SERVER_PORT`            | Port for the API server.                          | `8080`                   |
| `SERVER_HOST`            | Host for the API server.                          | `0.0.0.0`                |
| `STORAGE_PATH`           | Path to store data.                               | `/app/data`              |
| `REPO_PATH`              | Path to store cloned Git repositories.            | `/app/data/repos`        |
| `BINARY_PATH`            | Path to store compiled binaries.                  | `/app/data/binaries`     |
| `ADMIN_TOKEN`            | Secret token for accessing admin endpoints.       | `change-me-in-production`|
| `API_KEYS_ENABLED`       | Enable or disable API key authentication.         | `true`                   |
| `EXECUTOR_MAX_CONCURRENT`| Maximum number of concurrent executions.          | `10`                     |
| `EXECUTOR_TIMEOUT`       | Default execution timeout.                        | `5m`                     |
| `EXECUTOR_MAX_MEMORY_MB` | Maximum memory for each execution (not yet implemented). | `512`                    |

##  API Usage

The API is documented using OpenAPI (Swagger). You can access the interactive documentation at `http://localhost:8080/api/v1/docs`.

### Endpoints

-   `GET /api/v1/health`: Health check.
-   `GET /api/v1/docs`: Swagger UI.
-   `GET /api/v1/openapi.json`: OpenAPI specification.

#### Binary Management (`/api/v1/binaries`)

-   `GET /`: List all binaries.
-   `POST /`: Create a new binary from a Git repository.
-   `GET /{id}`: Get details of a binary.
-   `PUT /{id}`: Update a binary's configuration.
-   `DELETE /{id}`: Delete a binary.
-   `POST /{id}/build`: Build a binary.

#### Execution (`/api/v1/execute`)

-   `POST /`: Execute a binary.
-   `GET /{id}`: Get the status and output of an execution.
-   `DELETE /{id}`: Stop a running execution.

## 🛠️ Development

For development, you can use the provided `Makefile` for common tasks.

### Makefile Commands

-   `make build`: Build the Go binary.
-   `make run`: Run the application locally.
-   `make test`: Run tests.
-   `make clean`: Clean up build artifacts.
-   `make docker-build`: Build the Docker image.
-   `make docker-run`: Run the application in a Docker container.

## 🧪 Testing

The project uses Go's built-in testing framework. Unit tests are co-located with the code they test.

### Coverage Report

Overall Test Coverage: **39.5%**

| Package                     | Coverage |
| :-------------------------- | :------- |
| `go_runner/cmd/go_runner`   | 0.0%     |
| `go_runner/internal/api`    | 68.7%    |
| `go_runner/internal/config` | 0.0%     |
| `go_runner/internal/executor` | 84.9%    |
| `go_runner/internal/models` | No Test Files |
| `go_runner/internal/repository` | 0.0%     |
| `go_runner/internal/storage` | 0.0%     |

To run tests and generate a coverage report:

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
