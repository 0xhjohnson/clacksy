package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/jackc/pgtype"
	pgtypeuuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	templateCache map[string]*template.Template
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	addr := ":" + port

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	dbpool, err := openDbPool(databaseURL)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer dbpool.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		templateCache: templateCache,
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  2 * time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

func openDbPool(dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.ConnInfo().RegisterDataType(pgtype.DataType{
			Value: &pgtypeuuid.UUID{},
			Name:  "uuid",
			OID:   pgtype.UUIDOID,
		})
		return nil
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return pool, nil
}
