package pgbouncer

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/infobloxopen/dsnutil/pg"
)

var (
	//go:embed pgbouncer.go.tpl
	pgbouncerTemplateContent string
	pgbouncerTemplate        = template.Must(template.New("pgbouncer").Parse(pgbouncerTemplateContent))
)

type database struct {
	Name     string
	Host     string
	Port     int
	Username string
	Password string
}

type config struct {
	Databases               []database
	ListenPort              int
	ListenAddr              string
	AuthType                string
	AuthFile                string
	LogFile                 string
	PidFile                 string
	AdminUsers              []string
	RemoteUserOverride      string
	RemoteDBOverride        string
	IgnoreStartupParameters []string
	atomicWrite             func(string, string) error
}

func (c *config) toIni(filename string) error {
	buf := &bytes.Buffer{}
	if err := pgbouncerTemplate.Execute(buf, c); err != nil {
		return err
	}
	if c.atomicWrite == nil {
		c.atomicWrite = atomicWrite
	}
	return c.atomicWrite(filename, buf.String())
}

func (c *config) toUsersIni(filename string) error {
	buf := &bytes.Buffer{}
	for _, db := range c.Databases {
		if _, err := fmt.Fprintf(buf, "%s %s\n", db.Username, db.Password); err != nil {
			return fmt.Errorf("failed to write user: %w", err)
		}
	}
	if c.atomicWrite == nil {
		c.atomicWrite = atomicWrite
	}
	return c.atomicWrite(filename, buf.String())
}

// dbOptions is type that validates db options passed as a map.
type dbOptions map[string]string

func (opts dbOptions) Validate() error {
	if opts["host"] == "" {
		return errors.New("host value not found in db credential")
	}

	if opts["port"] == "" {
		return errors.New("port value not found in db credential")
	}

	if opts["dbname"] == "" {
		return errors.New("dbname value not found in db credential")
	}

	if opts["user"] == "" {
		return errors.New("user value not found in db credential")
	}

	if opts["password"] == "" {
		return errors.New("password value not found in db credential")
	}
	return nil
}

func (opts dbOptions) GetHost() string {
	return opts["host"]
}

func (opts dbOptions) GetPort() int {
	i, err := strconv.Atoi(opts["port"])
	if err != nil {
		return 5432
	}
	return i
}

func (opts dbOptions) GetDBName() string {
	return opts["dbname"]
}

func (opts dbOptions) GetUser() string {
	return opts["user"]
}

func (opts dbOptions) GetPassword() string {
	return opts["password"]
}

type configOption func(*config)

func parseDBCredentials(dsn string) (dbOptions, error) {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pg.ParseURL(dsn)
		if err != nil {
			return nil, err
		}
	}
	pgOptions, err := pg.ParseOpts(dsn)
	if err != nil {
		return nil, err
	}
	o := dbOptions(pgOptions)
	return o, o.Validate()
}

func withDBOptions(dsns []string) configOption {
	return func(c *config) {
		for _, dsn := range dsns {
			opts, err := parseDBCredentials(dsn)
			if err != nil {
				panic(err)
			}
			c.Databases = append(c.Databases, database{
				Name:     opts.GetDBName(),
				Host:     opts.GetHost(),
				Port:     opts.GetPort(),
				Username: opts.GetUser(),
				Password: opts.GetPassword(),
			})
		}
	}
}

func newConfig(option ...configOption) *config {
	c := &config{
		Databases: make([]database, 0),
	}
	for _, opt := range option {
		opt(c)
	}
	return c
}
