# Books-Management-System

This project is a Go-based backend application that manages book-related operations, utilizing Kafka for event-driven messaging, Redis for caching, and PostgreSQL for data storage. The application is containerized using Docker, with Docker Compose orchestrating services like Kafka, Redis, and PostgreSQL to ensure smooth communication between them. Swagger is integrated for API documentation, allowing easy interaction with the API endpoints. The entire application is deployed in AWS, leveraging the cloud platform for scalability and robust service management.


## Table of Contents

- [Prerequisites](#prerequisites)
- [Go Binary Building](#go-binary-building)
- [Install Kafka Redis Postgres](#install-kafka-redis-postgres)
- [Run the Go Binary](#run-the-go-binary)
- [Access the server and swagger](#access-the-server-and-swagger)

## Prerequisites

Before you begin, make sure you have the following software installed on your system:

- **Go** (for the backend development) [Download Go](https://go.dev/dl/)
- **Docker** (for containerization and orchestration) [Download Docker]
- **Docker** Compose (for managing multi-container Docker applications) [Install Docker Compose]

  
## Go Binary Building

### Step 1: Clone the Repository

Clone the repository to your local machine:

  ```
  git clone https://github.com/arepala-uml/books-management-system.git
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
  This will create the binary file `book-management-store` in the current directory.
      

## Install Kafka Redis Postgres

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

## Run the Go Binary
#### Step 1: Navigate to the Project Directory
  First, navigate to the books-management-system directory:
  ```
  cd books-management-system/
  ```

#### Step 2: Run the Binary in the Background Using `nohup`
  Now, use the nohup command to run the Go application in the background:
  ```
  nohup ./book-management-store > nohup.out 2>&1 &
  ```
  `nohup` ensures that the process will continue running in the background.

#### Step 3: Verify the Process is Running
  To verify if your Go application is running, use the following command:
  ```
  ps aux | grep book-management-store
  ```

#### Step 4: Check the `nohup.out` and `app.log`
  After running the above command, the output will be saved to nohup.out by default. 
  To verify the output, check the contents of the nohup.out file:
  ```
  cat nohup.out
  ```
#### Step 5: Monitor the Logs
  You can monitor the logs by checking the contents of the `nohup.out` or `app.log` files:
  ```
  tail -f nohup.out
  ```
  This command will display the latest logs in real time.
  Or if you want to monitor `app.log`:
  ```
  tail -f app.log
  ```

## Access the server and swagger.

#### Step 1: Access the Machine via IP Address
  To access your machine, you need to use the IP address or the public DNS of your server. 
  If you're running on an AWS EC2 instance, you can find the public IP address or public DNS 
  from the AWS Management Console under "Instances."

  Once you have the IP address, you can access the server like this:
  ```
  http://<your-server-ip>:<SERVER_PORT>/books/
  ```
  Replace `<your-server-ip>` with the public IP or DNS of your server.

  Replace `<SERVER_PORT>` with the port your Go application is running on port (9010)


#### Step 2: Swagger URL Access
  After navigating to the correct Swagger URL (/swagger), you should be able to see the Swagger UI and interact with the API documentation.
  ```
  http://<your-server-ip>:<SERVER_PORT>/swagger/
  ```
  This will allow you to view and test your API endpoints from the Swagger UI interface.

  Replace `<your-server-ip>` with the public IP or DNS of your server.

  Replace `<SERVER_PORT>` with the port your Go application is running on port (9010)

