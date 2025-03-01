package main

import (
	"flag"
	"fmt"
	"go-stripe/internal/driver"
	"go-stripe/internal/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
	secretkey string
	frontend  string
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
	DB       models.DBModel
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf(fmt.Sprintf("Starting Back End server in %s mode on port %d", app.config.env, app.config.port))

	return srv.ListenAndServe()
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4001, "Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application environment {development|production|maintenance}")
	flag.StringVar(&cfg.db.dsn, "dsn", "postgres://postgres:qwe@localhost:5432/widgets?TimeZone=Asia/Irkutsk&sslmode=disable", "DSN")
	flag.StringVar(&cfg.smtp.host, "smtphost", "app.debugmail.io", "smtp host")
	flag.StringVar(&cfg.smtp.username, "smtpusername", "565dc7ab-4ccf-44d4-84a1-2d319d434a76", "smtp username")
	flag.StringVar(&cfg.smtp.password, "smtppassword", "13649237-2f23-4cb3-924d-ee630fd85820", "smtp password")
	flag.IntVar(&cfg.smtp.port, "smtpport", 25, "smtp port")
	flag.StringVar(&cfg.secretkey, "secret", "bRWmrwNUToNUuzckjxcFlHZjxHkjrzKP", "secret key")
	flag.StringVar(&cfg.frontend, "frontend", "http://localhost:4000", "url to front end")

	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")
	cfg.stripe.key = os.Getenv("STRIPE_KEY")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	conn, err := driver.OpenDB(cfg.db.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer conn.Close()

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
		DB:       models.DBModel{DB: conn},
	}

	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}
}
