package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/0xhjohnson/clacksy/models"
	"github.com/0xhjohnson/clacksy/sqlite"
	"github.com/0xhjohnson/clacksy/ui"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jackc/pgtype"
	pgtypeuuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	dsn  = flag.String("dsn", ":memory:", "datasource name")
	addr = flag.String("addr", ":8080", "bind address")
)

func main() {
	flag.Parse()

	databaseURL := os.Getenv("DATABASE_URL")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	if *dsn == "" {
		errorLog.Fatal(fmt.Errorf("dsn required"))
	} else if *addr == "" {
		errorLog.Fatal(fmt.Errorf("bind address required"))
	}

	dbpool, err := openDbPool(databaseURL)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer dbpool.Close()

	db := sqlite.NewDB(":memory:")
	err = db.Open()
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache(ui.Files)
	if err != nil {
		errorLog.Fatal(err)
	}

	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(dbpool)
	sessionManager.Lifetime = 12 * time.Hour

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(os.Getenv("B2_KEY_ID"), os.Getenv("B2_APP_KEY"), ""),
		Endpoint:         aws.String("s3.us-west-004.backblazeb2.com"),
		Region:           aws.String("us-west-004"),
		S3ForcePathStyle: aws.Bool(true),
	}
	sess := session.Must(session.NewSession(s3Config))
	s3Client := s3.New(sess)

	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		sessionManager: sessionManager,
		templateCache:  templateCache,
		users:          &models.UserModel{DB: dbpool},
		soundtests:     &models.SoundTestModel{DB: dbpool},
		parts:          &models.PartsModel{DB: dbpool},
		votes:          &models.VoteModel{DB: dbpool},
		s3Client:       s3Client,
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		IdleTimeout:  2 * time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDbPool(dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
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

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return pool, nil
}

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	sessionManager *scs.SessionManager
	templateCache  map[string]*template.Template
	users          *models.UserModel
	soundtests     *models.SoundTestModel
	parts          *models.PartsModel
	votes          *models.VoteModel
	s3Client       *s3.S3
}
