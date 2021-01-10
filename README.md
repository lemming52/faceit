# Faceit Tech Test

## Installation Prerequisites

The following tools are required to run the service locally:
* Docker / Docker compose
* The AWS CLI, with a local dummy configuration (`aws configure --profile localstack
` with dummy values for keys)
* Golang
* [Localstack](https://github.com/localstack)

For full development, [Redoc](https://github.com/Redocly/redoc) is also used to generate the static doc page.

## Running

### Service

The service runs locally using a localstack container to emulate the database and messaging implementations, while the service container is a separate container managed in the same docker-compose file.

```
# Build the image
docker build -t faceit .

# run docker-compose to start services
docker-compose up

# Construct the databases, messaging and populate some test entries
source local-stack.sh
```

### Usage

The included postman collection has the set of endpoints for the service, and the full docs are in the `swagger.yaml` file and on the docs endpoint, but the headlines are below (the service runs on `localhost:3000`):

URL | Method | Description
----|--------|------------
`/healthcheck` | Get | Display basic service status info
`/docs` | Get | Display the pre-render HTML docs
`/users` | Get | Filter users by provided query params
`/users` | Post | Add a new user
`/users/{id}` | Get | Retrieve a specific user
`/users/{id}` | Put | Update a specific user
`/users/{id}` | Delete | Delete a specific user

### Unit tests

The unit tests of the handlers (with mocked interface clients) can be run by
```
make test
```

### Full Component Tests

The component tests here are a pseudo-substitute for testing on a canary/dev environment, but use the endpoints of a running service. To run these the service docker-compose needs to be up, running and populated.
```
make componenttests
```

Note that these will fail in two cases:
* You overwrite details of the test users
* You add more users, taking the total count up. This will only fail the count test for obvious reasons

## Brief

Remark | Comment
-------|--------
Each user entity must consist of a first name, last name, nickname, password, email and country. | Defined in [model](https://github.com/lemming52/faceit/blob/master/model/model.go#L11)
The service must allow you to ... | Each operation has a distinct endpoint
A sensible storage mechanism for the Users / The ability to send events to notify other interested services of changes to User entities | The service code sees an interface object to handle data access and messaging, both of these are provided by localstack AWS components, DynamoDB and SNS.
Meaningful logs | Logging messages thoughout with exposed fields where required
Self-documenting end points | RESTful design of the users endpoint and rendered docs hosted on `/docs`
Health checks | Healthcheck endpoint built into service

# Discussion

The sections below deal with more loose discussion on the decisions, extensions and assumptions.

## Assumptions

* Users are uniquely identified by the generated key, subsequent services will use that internal ID always when referring to a user.

* The input is basic alphanumeric for names and entries; not requiring full unicode, RTL or other non latin character sets

* *It is sufficient to monitor only the API service / The DB & Messaging services can be trusted* - The healthcheck only reports on the API, it assumes the DB and the messaging is connected correctly and running correctly. This solution assumes that these services are reliable and monitoring of them is not the responsibility of this particular service.

* *I have not considered error recovery, crash handling or operation conflict resolution at scale* - The exact operational behaviour when the service goes down depends on user desire, but this service is built on the assumption that the DB and messaging are stable, any error case can be returned to the user as an error, and there's no need to manage graceful shutdown and restart; if this service fails, it should just be started again.

* *The input to this service is controlled.* - I feel this is important as it relates to some of the checks we need to do in the service for safety in a production environment. My assumption is that the form data to add or update a user is coming from some authenticated app, say in a webpage, so I don't need to produce safety checks for the brief. Specifically I'm referring to simple things like ensuring the country code is valid (dropdown on form), ensuring the nickname contains allowed characters or more serious things like ensuring any externally entered data is cleaned.

* *Handling of sensitive data and PII is outwith this tests scope.* - In my example email and password will be stored unencrypted in the DB. In a production environment I'd also make sure that the request bodies for these operations are not logged, so as to not expose the sensitive data.

* Email collisions are fine. - More a simplicity thing for time rather than a difficulty, I mention implementing this in the extensions section.

* Filter/Search functionality is less prioritised than the act to storing and managing user lifecycles. - I used DynamoDB, partly as I'm familiar with it, but also as in terms of a DB for storing specific structures scalably and reliably it's a good choice. Where it's less strong is on the searchability; fuzzy search or things like that are trickier and can get expensive.

* Filter/Search functionality is exact match. - Relates to the point below, but for the sake of the brief assumed recovering exact values was needed.


## Alternatives

I've built this using AWS tooling, as their services fit the bill of the brief, but partly as usage of localstack made local emulation simple and it's where I have experience. Also, the code itself uses clients so the business logic of the user app is isolated from such service decisions.

There are some obvious possible alternatives. Firstly GCP has components for both SNS and Dynamo in Pub/Sub and Firestore. My experience with GCP is that they're better for small projects, but firestore runs into limitations, such as it indexes all columns, which may not be desirable.

DB alternatives are MongoDB or Elasticsearch, especially if the filter/search functionality is more important. Equally a more conventional SQL database could be used; I chose noSQL here as to the scalability and the flexibility benefits.

I have less experience with alternative messaging systems, but the big options are Kafka or RabbitMQ, if you wanted to do something more than just the SNS style notifications.

## Extensions

### Production readiness

As discussed in the brief, this was not required to be production ready. To do so the clients, dynamo and sns, would need rewriting to the full production configuration and the checks i mention throughout the readme applied.

### Rollbacks, uniqueness checks, allowed parameters

There's a few things that sprang to mind which I didn't implement in my proof of concept. As an example, if the message publication fails post user creation; do we want to keep the user or consider that a failure, do we add retries? For the exercise I ploughed on, but it's straightforward to add a finally style block that in the case of any errors restores the prior situation, or/and add retry code.

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

So, in a production environment, the component tests would be completely seperate from the service code, would seed and tidy up entities and messages they need.

### Docs

There's a few different options for linking the swagger generation of docs I've used here to the actual code (https://github.com/swaggo/swag, https://github.com/go-swagger/go-swagger), but I've not used any of them prior to this. In this case the swagger file will be manually maintained and compiled to an HTML file which is then added to a repo. You can shift the order of operations around with this, the HTML could be generated during the dockerised build in production rather than manually here.

## Clients

For both the clients, since they're used here primarily for the tests, I've not spent much time preparing them for a live environment; the regions, credentials and aws constants like table name and such are hardcoded into the instantiation methods. In a production environment, such details (which I would assume would also relate to which environment out of dev, prod...) would be passed to the execution container outwith the code and read from the environment, rather than being placed in the code.

## Automated Testing  / Healthchecks

This service healthcheck is designed to be repeatedly called by some exteral automation. The healthcheck could be expanded to perform some basic operations to ensure integrity rather than just responding, or the external automation could include calls to the other endpoints with expected values.