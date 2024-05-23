# RabbitMQ message queue and Redis caching with Golang

**RabbitMQ** implements a messaging protocol called Advanced Message Queueing Protocol (AMQP). 
It uses it to support worker queues. It also supports many other data exchange patterns, such as the following:
- Publish/Subscribe
- Topic/Subscription
- Routing messages
- Remote Procedure Call (RPC)

**Redis** is a wonderful open source solution for caching high-read configuration/information. 
It is a key/value pair store and has faster reads thanks to its in-memory storage. 
An example of a key/value pair store is a media website where a few articles are set fixed on their home page for a few hours.

This repo build a system that can handle requests for the following scenarios:
- The server should save information to the database as one operation.
- It should send an email to the given email address.
- It should perform a long-running job and POST the result to a callback. This is known as a web-hook.

Let's say these three operations are asynchronous and long-running. 
We need a mechanism to facilitate a long-running process that has the following characteristics:
- The client can fire an API and receive a job ID back.
- The job is pushed onto a queue with the respective message format.
- A worker picks the job and starts performing it.
- Finally, the worker saves the result on various endpoints and sends the status to the database.

There are 3 tasks adding to the project:
1. Add a new route to collect the job ID from a client => /job/status
2. Add a new handler to fetch the job status from Redis => statusHandler()
3. Whenever someone adds a new job, we need to write status to Redis in every stage of the job life cycle
   - w.redisClient.Set(job.ID.String(), "STARTED", 0)
   - w.redisClient.Set(job.ID.String(), "IN PROGRESS", 0)
   - w.redisClient.Set(job.ID.String(), "DONE", 0)
