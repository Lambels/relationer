# Relationer
This project provides a real-time interface to view highly-customisable relationships (graph relationships) between people. The relationer server produces messages which are sent to a centralised message broker to be consumend by you on either: the relationer cli recieving point, provided frontend, or your own customisable app.

# Get Started
**Using realtioner with golang? try the [golang client](https://github.com/Lambels/relationer/tree/main/client)**
## Binary Packages
All the final binaries are under the `./cmd` folder:
- `./cmd/relationer-server` - the relationer server binary
- `./cmd/relationer` - the relationer cli binary
### Relationer Server Binary:
Flags:
```
Usage of relationer-server:
  -backup
        use a backup datastore (default true)
  -bk-addr string
        address of the broker (default "amqp://guest:guest@localhost:5672")
  -cache-addr string
        address of the cache (default "localhost:6379")
  -db-addr string
        dns of the database (default "postgres://username:password@localhost/relationer?sslmode=disable")
  -serv-addr string
        address of the server (default ":8080")
```
The relationer server can start with or without a persistent data-store (postgres) controlled via the `-backup` flag. When true the `-db-addr` flag will be used to connect to the postgres instance where any previous data will be loaded on the new relationer server and any new data will be added to the postgres instance.
### Relationer Cli Binary:
Usage:
```
USAGE
  relationer is a cli tool to interact with the relationer server

SUBCOMMANDS
  add-person      Create a user (node)
  add-friendship  Create a friendship (edge) uni-directional from id1 -> id2
  get-depth       Get depth between 2 nodes
  get-friendship  Get the relationships of a person
  get-person      Get the person with provided id
  listen          Listen for events

FLAGS
  -p http://localhost:8080  api enpoint for relationer
  -v=false                  log verbose output
```

# Quickstart
> Must already have postgres instance running

> Must already have redis instance running

> Must already have rabbitmq instance running

> Have `$GOPATH` in yout path

Build the relationer project by cloning the repo then run:
```
$ go install ./cmd/...
```

## Relationer Server (persistent store)
```
$ relationer-server -bk-addr rabbitmq-url -cache-addr redis-url -db-addr postgres-dsn
2022/05/22 13:02:11 server starting on :8080 ...
```
To kill the server `Ctrl-C`

## Relationer Cli
If you have a relationer server running, you can now interface with it via the relationer cli binary and all its sub commands.
### Create two people and have them linked via a mutual friendship:
```
$ relationer -v add-person Lambels
created person with id 1 at 0000-00-00 00:00:00 +0000 UTC OK
Process took x ms
$ relationer -v add-person Lambili
created person with id 2 at 0000-00-00 00:00:00 +0000 UTC OK
Process took x ms
$ relationer -v add-friendship -mutual 1 2
created friendship between 1 and 2 OK
created friendship between 2 and 1 OK
Process took x ms
```
### Start an event listener:
The relationer cli event listener listens to messages from the relationer server.
You can specify the events you want to listen to via: (rabbitmq routing keys)
- `relationer -v listen person.created` listen for created persons
- `relationer -v listen person.deleted` listen for deleted persons
- `relationer -v listen friendship.created` listen for created friendships
- `relationer -v listen person.created person.deleted` listen for created or deleted persons
- `relationer -v listen -all` listen for all events ("#" routing key)
  
Example running in the background of the commands above:
```
$ relationer -v listen -all
(re)dial to amqp://guest:guest@localhost:5672
(re)dial to amqp://guest:guest@localhost:5672
subscribed and ready to recieve...
[New Person] Id: 1 | Name: Lambels | Created At: 0000-00-00 00:00:00 +0000 UTC
[New Person] Id: 2 | Name: Lambili | Created At: 0000-00-00 00:00:00 +0000 UTC
[New Friendship] Person 1: 1 with Person 2: 2
[New Friendship] Person 1: 2 with Person 2: 1
```

**Commands: cli -> relationer and server -> relationer-server**
![demo giphy](demo.gif)
