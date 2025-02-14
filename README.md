# Books-Management-System

This project is a Go-based backend application that manages book-related operations, utilizing Kafka for event-driven messaging, Redis for caching, and PostgreSQL for data storage. The application is containerized using Docker, with Docker Compose orchestrating services like Kafka, Redis, and PostgreSQL to ensure smooth communication between them. Swagger is integrated for API documentation, allowing easy interaction with the API endpoints. The entire application is deployed in AWS, leveraging the cloud platform for scalability and robust service management.


## Table of Contents

- [Prerequisites](#prerequisites)
- [Go Binary Building](#go-binary-building)
- [Install Kafka, Redis, PostgreSQL](#installation-kafka-redis-postgresql)

## Prerequisites

Before you begin, make sure you have the following software installed on your system:

- **Go** (for the backend development) [Download Go](https://go.dev/dl/)
- **Docker** (for containerization and orchestration) [Download Docker]
- **Docker** Compose (for managing multi-container Docker applications) [Install Docker Compose]

  
## Go Binary Building

### Step 1: Clone the Repository

Clone the repository to your local machine:

  ```
  git https://github.com/arepala-uml/books-management-system.git
  ```
Move to the project directory:
  ```
  cd books-management-system/
  ```

### Step 2: Cross-Compiling Go Binary for Different Operating Systems
  * On MacOS:
    ```
    GOOS=darwin GOARCH=amd64 go build -o book-management-store main.go
    ```
  * On Linux:
    ```
    GOOS=linux GOARCH=amd64 go build -o book-management-store main.go
    ```
  * On Windows:
    ```
    GOOS=windows GOARCH=amd64 go build -o book-management-store main.go
    ```
      

## Setting Up Kafka, Redis, and PostgreSQL Services

#### Step 1: Navigate to Frontend Folder
  Navigate to the project directory
  ```
  cd books-management-system/
  ```

#### Step 2: Bring up all services
  Run the following command to start all the services (Zookeeper, Kafka, Redis, PostgreSQL):
  ```
  docker-compose up -d
  ```
  This will bring up all the services in the background.

#### Step 3: Verify running services
  To check the status of all services and ensure they are running:
  ```
  docker-compose ps
  ```
  This will show the list of services and their current status (running, exited, etc.).

#### Step 4: View logs of specific services
  To check the logs of any specific service, use the following command:
  ```
  docker-compose logs <service-name>
  ```

#### Step 5: Check specific service ports
  You can check the following services and their exposed ports:
  1. Kafka: `localhost:29092`
  2. Redis: `localhost:6379`
  3. PostgreSQL: `localhost:5432`
     
  You can connect to each of these services from your local machine using these ports.

#### Step 6: Stop the services
  To stop all services:
  ```
  docker-compose down
  ```
  This will stop all running services and remove their containers.

## Installation Using Docker (optional)



