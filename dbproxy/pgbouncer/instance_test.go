package pgbouncer

import (
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/kylelemons/godebug/diff"
)

func TestWithPGBouncerExecutePath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default",
			args: args{
				// path: "pgbouncer",
			},
			want: "pgbouncer",
		},
		{
			name: "with foo",
			args: args{
				path: "foo",
			},
			want: "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := New(WithPGBouncerExecutePath(tt.args.path))
			if got := i.daemonPath; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithPGBouncerExecutePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithDSN(t *testing.T) {

	tests := []struct {
		name string
		args []InstanceOption
		want map[string]string
	}{
		{
			name: "default",
			args: []InstanceOption{
				WithDSN("foo", ""),
			},
			want: map[string]string{"foo": ""},
		},
		{
			name: "foo",
			args: []InstanceOption{
				WithDSN("foo", "foo"),
			},
			want: map[string]string{"foo": "foo"},
		},
		{
			name: "foo, bar",
			args: []InstanceOption{
				WithDSN("foo", "foo"),
				WithDSN("bar", "bar"),
			},
			want: map[string]string{"foo": "foo", "bar": "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := New(tt.args...)
			if got := i.dsns; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}

// withDaemonArgs sets the arguments for the pgbouncer instance for testing.
func withDaemonArgs(args ...string) InstanceOption {
	return func(i *Instance) {
		i.daemonArgs = args
	}
}

func TestInstance_Start(t *testing.T) {

	tests := []struct {
		name       string
		arg        *Instance
		wantOnExit error
		wantErr    bool
	}{
		{
			name:    "not find executeable",
			arg:     New(WithPGBouncerExecutePath("!gonnafindit!")),
			wantErr: true,
		},
		{
			name:    "no dsn set",
			arg:     New(WithPGBouncerExecutePath("sleep"), withDaemonArgs("0.1")),
			wantErr: true,
		},
		{
			name: "sleep 0.1",
			arg: New(
				WithPGBouncerExecutePath("sleep"),
				withDaemonArgs("0.1"),
				WithDSN("foo", "postgres://foo:bar@remote-db:5432/foo?sslmode=require"),
			),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOnExit, err := tt.arg.Start()
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Instance.Start() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			} else {
				if tt.wantErr {
					t.Errorf("Instance.Start() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
			if tt.wantOnExit != nil {
				select {
				case got, ok := <-gotOnExit:
					if !ok {
						t.Errorf("Instance.Start() got closed channel")
						return
					} else {
						if !strings.Contains(got.Error(), tt.wantOnExit.Error()) {
							t.Errorf("Instance.Start() = %v, want %v", got, tt.wantOnExit)
							return
						}
					}
				case <-time.After(1 * time.Second):
					t.Errorf("Instance.Start() got timeout %v", tt.arg)
					tt.arg.Stop()
					return
				}
			}
		})
	}
}

type testableAtomicWrite map[string]string

func (t testableAtomicWrite) atomicWrite(pth, data string) error {
	t[pth] = data
	return nil
}

func TestInstance_generateConfigs(t *testing.T) {
	type fields struct {
		dsns      map[string]string
		localAddr string
	}
	tests := []struct {
		name        string
		fields      fields
		wantErr     bool
		wantContent map[string]string
	}{
		{
			name: "default",
			fields: fields{
				dsns: map[string]string{
					"foo": "postgres://foo:bar@remote-db:5432/foo?sslmode=require",
				},
				localAddr: "127.0.0.1:5432",
			},
			wantErr: false,
			wantContent: map[string]string{
				"pgbouncer.ini": `[databases]
foo = host=remote-db port=5432 user=foo dbname=foo

[pgbouncer]
listen_port = 5432
listen_addr = 127.0.0.1
admin_users = []
auth_type = trust
auth_file = userlist.txt
ignore_startup_parameters = extra_float_digits
client_tls_sslmode = require
client_tls_key_file=dbproxy-client.key
client_tls_cert_file=dbproxy-client.crt
server_tls_sslmode = require`,
				"userslist.txt": "foo bar\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// don't create temp files, is messes up our comparisons
			aw := make(testableAtomicWrite)

			// mock our instance
			i := &Instance{
				dsns:          tt.fields.dsns,
				localAddr:     tt.fields.localAddr,
				atomicWrite:   aw.atomicWrite,
				iniFilename:   "pgbouncer.ini",
				usersFilename: "userslist.txt",
			}
			if err := i.generateConfigs(); (err != nil) != tt.wantErr {
				t.Errorf("Instance.generateConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// break down this comparison so its easier to debug
			// first check if the length of the maps are the same
			if len(tt.wantContent) != len(aw) {
				t.Errorf("Instance.generateConfigs() mismatch len(want) != len(got)")
				t.Errorf("Instance.generateConfigs() got = %v, want %v", map[string]string(aw), tt.wantContent)
				return
			}
			// extract keys and sort them to compare got and want
			var gotKeys []string
			for k := range aw {
				gotKeys = append(gotKeys, k)
			}
			sort.Strings(gotKeys)
			var wantKeys []string
			for k := range tt.wantContent {
				wantKeys = append(wantKeys, k)
			}
			sort.Strings(wantKeys)
			if cdiff := cmp.Diff(wantKeys, gotKeys); cdiff != "" {
				t.Errorf("Instance.generateConfigs() keys mismatch (-want +got):\n%s", cdiff)
				return
			}
			// do item by item comparison of got and want
			for _, k := range gotKeys {
				if cdiff := cmp.Diff(tt.wantContent[k], aw[k]); cdiff != "" {
					d := diff.Diff(tt.wantContent[k], aw[k])
					t.Errorf("Instance.generateConfigs() mismatch (-want +got):\n%s", d)
					return
				}
			}
		})
	}
}
