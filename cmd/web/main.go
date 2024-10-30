package main

import (
	"cosmasgithinji.net/simplesnippetbox/pkg/models"
	"cosmasgithinji.net/simplesnippetbox/pkg/models/mysql"
	"crypto/tls"
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
	"github.com/joho/godotenv"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

type contextKey string

var contextKeyUser = contextKey("user")

// struct to hold app-wide dependencies
//
//	type application struct {
//		errorLog      *log.Logger
//		infoLog       *log.Logger
//		session       *sessions.Session
//		snippets      *mysql.SnippetModel
//		templateCache map[string]*template.Template
//		users         *mysql.UserModel
//	}
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	session  *sessions.Session
	snippets interface {
		Insert(string, string, string) (int, error)
		Get(int) (*models.Snippet, error)
		Latest() ([]*models.Snippet, error)
	}
	templateCache map[string]*template.Template
	users         interface {
		Insert(string, string, string) error
		Authenticate(string, string) (int, error)
		Get(int) (*models.User, error)
	}
}

// provide a command line flag or .env file to run
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//flag.String() returns a pointer
	addrFlag := flag.String("addr", "", "HTTP Network Address")                          // --addr=:<port_number>
	dsnFlag := flag.String("dsn", "", "MySQL data source name")                          // --dsn=<user>:<db>@/<db>?parseTime=true
	secretFlag := flag.String("secret", "", "32-byte Secret Key for Session Management") // --secret=<secret_string>
	flag.Parse()

	addr, err := getFlagOrEnv(addrFlag, "ADDR")
	if err != nil {
		log.Fatal(err)
	}

	dsn, err := getFlagOrEnv(dsnFlag, "DSN")
	if err != nil {
		log.Fatal(err)
	}

	secret, err := getFlagOrEnv(secretFlag, "SECRET")
	if err != nil {
		log.Fatal(err)
	}

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)                  //INFO logger: destination, prefix, flags for additional info
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile) //INFO logger: destination, prefix, flags for additional info

	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html")
	if err != nil {
		errorLog.Fatal(err)
	}

	session := sessions.New([]byte(secret)) // init new session with secret
	session.Lifetime = 12 * time.Hour       // expire after 12 hours
	session.Secure = true

	app := &application{ //init a new instance of application
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
		snippets:      &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
		users:         &mysql.UserModel{DB: db},
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(), //call the app.routes() method
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server at %s", addr) //addr value from flag.String() is a pointer to the value and not the value itself
	// err := http.ListenAndServe(*addr, mux) //start a new web server
	// err = srv.ListenAndServe()
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

// wrap sql.Open and return sql.DB connection pool for given dsn
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
