# PressAndPlay
A web application to book slot in nearby sporting centers

## Goal:
We built an application with which a user can search for and book a court for various sports in public sporting centers across the city. The application will have 2 interfaces to interact, one for the user and another for the sporting center manager. When a user books a court, the respective sporting center manager will be notified and can approve or reject the appointment. Customers can provide feedback/ratings after the scheduled appointment, which will be used to order sporting courts on the catalog.

## Software Components:
● Frontend technologies - Angular
● Backend technologies
  ○ RPC / API interfaces - Using RESTful API and gRPC for user interaction and synchronous inter-service communication
  ○ Message queues - Using Kafka for events, notifications, and asynchronous inter-service communication.
  ○ Key-Value store - Using Redis to maintain user session details.
  ○ Message marshaling/encoding - Using Protobuf to marshall messages
  between services.
  ○ Containers - Using Docker to containerize/package each service.
  ○ SQL Databases - User and transactions data will be stored in PostgresSQL
  ○ NoSQL Databases - Courts data will be stored in MongoDB
  ○ AWS S3 - to store blob and file data
● Deployment - Kubernetes

### Demo: https://www.youtube.com/watch?v=fCsIyLKu1As

