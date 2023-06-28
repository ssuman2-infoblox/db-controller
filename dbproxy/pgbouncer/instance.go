package pgbouncer

import (
	"errors"
	"fmt"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Instance is a pgbouncer instance.
type Instance struct {
	dsns          map[string]string
	daemonPath    string
	daemonArgs    []string
	localAddr     string
	l             sync.RWMutex
	cmd           execCmd
	atomicWrite   func(path, data string) error // this is mockable for testing
	iniFilename   string
	usersFilename string
}

// New creates a new pgbouncer instance.
func New(options ...InstanceOption) *Instance {
	i := &Instance{
		dsns:        make(map[string]string),
		daemonPath:  "pgbouncer",
		daemonArgs:  []string{"-v", "-R"},
		localAddr:   "127.0.0.1:5432",
		atomicWrite: atomicWrite,
	}
	for _, option := range options {
		option(i)
	}
	return i
}

// WithListenAddr sets the listen address for the pgbouncer instance.
func WithListenAddr(addr string) InstanceOption {
	return func(i *Instance) {
		i.localAddr = addr
	}
}

// WithPGBouncerExecutePath sets the path to the pgbouncer executable.
func WithPGBouncerExecutePath(path string) InstanceOption {
	return func(i *Instance) {
		if path == "" {
			log.Warn("pgbouncer executable path is empty, using default")
			return
		}
		i.daemonPath = path
	}
}

// WithDSN sets the dsn for the pgbouncer instance.
func WithDSN(name, dsn string) InstanceOption {
	return func(i *Instance) {
		i.dsns[name] = dsn
	}
}

// InstanceOption is a functional option for pgbouncer instance.
type InstanceOption func(*Instance)

func (i *Instance) createTempConfigFilesAndClose() (err error) {

	f, err := os.CreateTemp("", "pgbouncer.ini.")
	if err != nil {
		return err
	}
	iniFilename := f.Name()
	f.Close()
	f, err = os.CreateTemp("", "pgbouncer.users.")
	if err != nil {
		return err
	}
	usersFilename := f.Name()
	f.Close()
	i.usersFilename = usersFilename
	i.iniFilename = iniFilename

	return nil
}

// Start starts the pgbouncer instance.
func (i *Instance) Start() (onExit chan error, err error) {
	i.l.Lock()
	defer i.l.Unlock()
	if i.cmd != nil {
		return nil, errors.New("instance already started")
	}
	if _, err := execLookPath(i.daemonPath); err != nil {
		return nil, err
	}

	if err := i.createTempConfigFilesAndClose(); err != nil {
		return nil, err
	}

	if err := i.generateConfigs(); err != nil {
		return nil, err
	}

	// create a tempfile for the config file
	args := append(i.daemonArgs, i.iniFilename)

	i.cmd = newExecCommand(i.daemonPath, args...)
	i.cmd.SetStderr(&pgbouncerLogger{})
	i.cmd.SetStdout(&pgbouncerLogger{})
	if err := i.cmd.Start(); err != nil {
		return nil, err
	}
	onExit = make(chan error, 1)
	started := make(chan struct{})
	go func() {
		close(started)
		onExit <- i.cmd.Wait()
		close(onExit)
		i.l.Lock()
		i.cmd = nil
		i.l.Unlock()
	}()
	<-started // wait till goroutine is started
	return onExit, nil
}

func (i *Instance) Stop() error {
	i.l.Lock()
	defer i.l.Unlock()
	if i.cmd == nil {
		return errors.New("instance not started")
	}
	if err := i.cmd.Kill(); err != nil {
		return err
	}
	i.cmd = nil
	return nil
}

func (i *Instance) Reload() error {
	i.l.Lock()
	defer i.l.Unlock()
	if i.cmd == nil {
		return errors.New("instance not started")
	}
	if err := i.cmd.Sighup(); err != nil {
		return fmt.Errorf("failed to reload pgbouncer: %s", err)
	}
	return nil
}

func (i *Instance) LoadDSN(name, dsn string) error {
	i.l.Lock()
	i.dsns[name] = dsn
	defer i.l.Unlock()

	if err := i.generateConfigs(); err != nil {
		return err
	}
	return i.Reload()
}

func (i *Instance) generateConfigs() error {
	if len(i.dsns) == 0 {
		return errors.New("no dsn set")
	}
	c, err := i.toConfig()
	if err != nil {
		return err
	}
	if i.atomicWrite != nil {
		c.atomicWrite = i.atomicWrite
	}

	if err := c.toIni(i.iniFilename); err != nil {
		return err
	}
	if err := c.toUsersIni(i.usersFilename); err != nil {
		return err
	}
	return nil
}

func (i *Instance) toConfig() (*config, error) {
	c := &config{}
	for name, dsn := range i.dsns {
		d, err := parseDBCredentials(dsn)
		if err != nil {
			return nil, err
		}
		c.ListenAddr, c.ListenPort, err = parseHostPort(i.localAddr)
		if err != nil {
			return nil, err
		}
		c.Databases = append(c.Databases, database{
			Name:     name,
			Host:     d.GetHost(),
			Port:     d.GetPort(),
			Username: d.GetUser(),
			Password: d.GetPassword(),
		})
	}
	return c, nil
}
