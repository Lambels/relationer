package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lambels/relationer/internal/graph"
	noop "github.com/Lambels/relationer/internal/no-op"
	"github.com/Lambels/relationer/internal/postgresql"
	"github.com/Lambels/relationer/internal/rabbitmq"
	"github.com/Lambels/relationer/internal/redis"
	"github.com/Lambels/relationer/internal/rest"
	"github.com/Lambels/relationer/internal/service"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/streadway/amqp"
)

type Config struct {
	databaseAddr string
	serverAddr   string
	cacheAddr    string
	brokerAddr   string
	backup       bool
	middleware   []func(http.Handler) http.Handler
	gStore       service.GraphStore
	store        service.Store
	broker       service.MessageBroker
}

func main() {
	conf := &Config{}

	// flags.
	flag.BoolVar(&conf.backup, "backup", true, "use a backup datastore")
	flag.StringVar(&conf.serverAddr, "serv-addr", ":8080", "address of the server")
	flag.StringVar(&conf.databaseAddr, "db-addr", "postgres://patrickarvatu:password@localhost/relationer?sslmode=disable", "dns of the database")
	flag.StringVar(&conf.brokerAddr, "bk-addr", "amqp://guest:guest@localhost:5672", "address of the broker")
	flag.StringVar(&conf.cacheAddr, "cache-addr", "localhost:6379", "address of the cache")
	flag.Parse()

	// surface lvl middleware.
	conf.middleware = append(
		conf.middleware,
		chimw.Logger,
		chimw.Recoverer,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()
	run(ctx, conf)
	log.Println("exiting ...")
}

func run(ctx context.Context, conf *Config) {
	// setup cache.
	cache := redis.NewCache(conf.cacheAddr)

	// setup stores and database.
	var store service.Store
	var gStore service.GraphStore
	var db *postgresql.DB
	if conf.backup {
		db = postgresql.NewDB(conf.databaseAddr)
		if err := db.Open(); err != nil {
			log.Printf("open database error: %v\n", err)
			return
		}

		store = postgresql.NewPostgresqlStore(db, cache)
		graph := graph.NewGraphStore(db.DB(), store, cache)
		if err := graph.Load(ctx); err != nil {
			log.Fatalf("%v\n", err)
		}
		gStore = graph
	} else {
		gStore = graph.NewGraphStore(nil, nil, cache)
		store = noop.NewNoopStore()
	}

	conf.store = store
	conf.gStore = gStore

	// setup rabbitmq.
	conn, err := amqp.Dial(conf.brokerAddr)
	if err != nil {
		log.Printf("dial amqp error: %v\n", err)
		return
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Printf("create channel amqp error: %v\n", err)
		return
	}

	if err := channel.ExchangeDeclare(
		"relationer", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // args
	); err != nil {
		log.Printf("declare exchange amqp error: %v\n", err)
		return
	}
	conf.broker = rabbitmq.NewRabbitMq(channel)

	srv, err := newServer(conf)
	if err != nil {
		log.Printf("server initialization error: %v\n", err)
		return
	}

	// run server in gorutine.
	go func() {
		log.Printf("server starting on %v ...", conf.serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server listen and serve error: %v\n", err)
		}
	}()

	<-ctx.Done()

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer func() {
		db.Close()
		channel.Close()
		conn.Close()
		cancel()
	}()

	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Printf("server shutdown error: %v\n", err)
	}
	log.Println("gracefull shutdown completed ...")
}

func newServer(conf *Config) (*http.Server, error) {
	router := chi.NewRouter()

	for _, mwFunc := range conf.middleware {
		router.Use(mwFunc)
	}

	rest.NewHandlerService(conf.gStore, conf.broker).RegisterRouter(router)

	// TODO: static routes

	return &http.Server{
		Handler:           router,
		Addr:              conf.serverAddr,
		ReadTimeout:       time.Second,
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		IdleTimeout:       time.Second,
	}, nil
}
