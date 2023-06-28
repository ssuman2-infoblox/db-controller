package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/infobloxopen/db-controller/dbproxy/pgbouncer"

	"github.com/infobloxopen/hotload/fsnotify"
)

func generatePGBouncerConfiguration(dsn string, port int, pbCredentialPath string) {
	dbc, err := pgbouncer.ParseDBCredentials(dsn)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	localHost := "127.0.0.1"
	localPort := port
	err = pgbouncer.WritePGBouncerConfig(pbCredentialPath, &pgbouncer.PGBouncerConfig{
		LocalDbName: dbc.GetDBName(), LocalHost: localHost, LocalPort: int16(localPort),
		RemoteHost: dbc.GetHost(), RemotePort: int16(dbc.GetPort()), UserName: dbc.GetUser(), Password: dbc.GetPassword()})
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func startPGBouncer() {
	ok, err := pgbouncer.Start()
	if !ok {
		log.Println(err)
		panic(err)
	}
}

func reloadPGBouncerConfiguration() {
	ok, err := pgbouncer.ReloadConfiguration()
	if !ok {
		log.Println(err)
		panic(err)
	}
}

func waitForDbCredentialFile(path string) {
	for {
		time.Sleep(time.Second)

		file, err := os.Open(path)
		if err != nil {
			log.Println("Waiting for file to appear:", path, ", error:", err)
			continue
		}

		stat, err := file.Stat()
		if err != nil {
			log.Println("failed stat file:", err)
			continue
		}

		if !stat.Mode().IsRegular() {
			log.Println("not a regular file")
			continue
		} else {
			break
		}
	}
}

var (
	dbCredentialPath string
	dbPasswordPath   string
	pbCredentialPath string
	port             int
)

func init() {
	flag.StringVar(&dbCredentialPath, "dbc", "./db-credential", "Location of the DB Credentials")
	flag.StringVar(&dbPasswordPath, "dbp", "./db-password", "Location of the unescaped DB Password")
	flag.StringVar(&pbCredentialPath, "pbc", "./pgbouncer.ini", "Location of the PGBouncer config file")
	flag.IntVar(&port, "port", 5432, "Port to listen on")
}

func main() {

	flag.Parse()

	if port < 1 || port > 65535 {
		log.Fatal("Invalid port number")
	}

	waitForDbCredentialFile(dbCredentialPath)
	s := fsnotify.NewStrategy()
	dsn, dsns, err := s.Watch(context.Background(), dbCredentialPath, nil)
	if err != nil {
		log.Fatal("NewWatcher failed: ", err)
	}

	// First time pgbouncer config generation and start
	generatePGBouncerConfiguration(dsn, port, pbCredentialPath)
	startPGBouncer()

	// Watch for ongoing changes and regenerate pgbouncer config

	go func() {

		for dsn := range dsns {
			if dsn == "" {
				log.Fatalf("could not get any more dsns, range closed")
			}
			// Regenerate pgbouncer configuration and signal pgbouncer to reload cconfiguration
			generatePGBouncerConfiguration(dsn, port, pbCredentialPath)
			reloadPGBouncerConfiguration()
		}

	}()
}
