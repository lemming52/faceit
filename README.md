# Faceit TT

docker / run docker compose

awscli installed

aws credentials may need configuring, but will be ignored


## Brief

###	A sensible storage mechanism for the Users

From the perspective of the service code, it needs some kind of persisent DB for the users.

### The ability to send events to notify other interested services of changes to User entities

To this end, any of the mutation endpoints (add, update, delete) needs to emit a message whenever a change is succesfully carried out. The others need not publish. For the sake of the exercise I've just configured a publisher attached to a queue for testing, and not done any sort of lifecycle configuration.

### Health checks

The service exposes a simple healthcheck endpoint that returns the service name and version. This can clearly be extended as required.

# TODO

Dockerise
Tidy logging to be consistent
Add index on email
Add tests
Add swagger renderer
Add check if exists

## Clients

For both the clients, since they're used here primarily for the tests, I've not spent much time preparing them for a live environment; the regions, credentials and aws constants like table name and such are hardcoded into the instantiation methods. In a production environment, such details (which I would assume would also relate to which environment out of dev, prod...) would be passed to the execution container outwith the code and read from the environment, rather than being placed in the code.

### DynamoDB

### SNS

## Assumptions

* *The input to this service is controlled.* - I feel this is important as it relates to some of the checks we need to do in the service for safety in a production environment. My assumption is that the form data to add or update a user is coming from some authenticated app, say in a webpage, so I don't need to produce safety checks for the brief. Specifically I'm referring to simple things like ensuring the country code is valid (dropdown on form), ensuring the nickname contains allowed characters or more serious things like ensuring any externally entered data is cleaned.

* *Handling of sensitive data and PII is outwith this tests scope.* - In my example email and password will be stored unencrypted in the DB. In a production environment I'd also make sure that the request bodies for these operations are not logged, so as to not expose the sensitive data.

* Email collisions are fine. - More a simplicity thing for time rather than a difficulty, I mention implementing this in the extensions section.

* Filter/Search functionality is exact match. - Relates to the point below, but for the sake of the brief assumed recovering exact values was needed.

* Filter/Search functionality is less prioritised than the act to storing and managing user lifecycles. - I used DynamoDB, partly as I'm familiar with it, but also as in terms of a DB for storing specific structures scalably and reliably it's a good choice. Where it's less strong is on the search ability; fuzzy search or things like that are trickier and can get expensive.

## Alternatives

I've built this using AWS tooling, as their services fit the bill of the brief, but partly as usage of localstack made local emulation simple and it's where I have experience. Also, the code itself uses clients so the business logic of the user app is isolated from such service decisions.

There are some obvious possible alternatives. Firstly GCP has components for both SNS and Dynamo in Pub/Sub and Firestore. My experience with GCP is that they're better for small projects, but firestore runs into limitations, such as it indexes all columns, which may not be desirable.

DB alternatives are MongoDB or Elasticsearch, especially if the filter/search functionality is more important. Equally a more conventional SQL database could be used; I chose noSQL here as to the scalability and the flexibility benefits.

I have less experience with alternative messaging systems, but the big options are Kafka or RabbitMQ, if you wanted to do something more than just the SNS style notifications.

## Extensions

### Rollbacks, uniqueness checks, allowed parameters

There's a few things that sprang to mind which I didn't implement in my proof of concept. As an example, if the message publication fails post user creation; do we want to keep the user or consider that a failure? For the exercise I ploughed on, but it's straightforward to add a finally style block that in the case of any errors restores the prior situation.

In addition, I debated a while about making the email address a secondary index. It seems a fair assumption to limit one account per email; to do so in the code you add a secondary index to the table, add checks on insertion if the address already exists and also on update, but I didn't do it for the test here.

### Update

Change the endpoint to not require all fields to be provided in request body, set criterion on which fields can be changed and which can't and so on.

### Searching

As discussed, this implementation is not designed to facilite fuzzy/soft user searches. If that was a requirement the usage of Dynamo would likely need to change, or you could work to develop something built on top of it.

### Middleware and shared tooling

Lots of the tools I've written smaller versions of here benefit from having tooling and middleware in established codebases. Shared baseline DynamoDB clients that are extended to fit the specific service mean that the core behaviour can be consistent across teams and configured to the teams particular infrastructure while allowing easy development of similar services.

Middleware can be added to the handler functions (as i've done here in a sense with the HandlerFunc) that adds functionality like authentication checks, logging and other every-request style operations.

Logging; here I've used logrus for simplicity, but both logging and errors benefit from consistent approachs and set constants. In addition i've avoided adding logging in the client implementations to avoid additional boilerplate and avoid double logging errors; the handler is the principle error record for this service. This could be more sophistcated. In addition in certain cases internal error messages are propogated to the request response, which may not be desireable.

TraceID; key to have, either attach to all new requests or propagate for tracing of requests between services.

### Component tests

I've added some light component tests; but the design of them is a little self referential. In an ideal world the seeding, checking and cleanup would be carried out by interacting directly with dynamo, but for the sake of reducing boiler plate I've just used endpoints to do that, and tried to order the endpoints in a way that builds up reliability.