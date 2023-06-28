package pgbouncer

import (
	"testing"
)

func TestDBOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "empty host",
			dsn:     "port=5432 dbname=bar user=u password=p",
			wantErr: true,
		},
		{
			name:    "ok",
			dsn:     "host=foo port=5432 dbname=bar user=u password=p",
			wantErr: false,
		},
		{
			name:    "no dbname",
			dsn:     "host=foo port=5432 user=u password=p",
			wantErr: true,
		},
		{
			name:    "no port",
			dsn:     "host=foo dbname=bar user=u password=p",
			wantErr: true,
		},
		{
			name:    "no user",
			dsn:     "host=foo port=5432 dbname=bar password=p",
			wantErr: true,
		},
		{
			name:    "no password",
			dsn:     "host=foo port=5432 dbname=bar user=u",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := parseDBCredentials(tt.dsn)
			if err := v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("DBOptions.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBOptions_Get(t *testing.T) {
	type val struct {
		Host     string
		Port     int
		DBName   string
		User     string
		Password string
	}
	tests := []struct {
		name string
		dsn  string
		want val
	}{
		// TODO: Add test cases.
		{
			name: "get options",
			dsn:  "host=foo port=5432 dbname=bar user=u password=p",
			want: val{
				Host:     "foo",
				Port:     5432,
				DBName:   "bar",
				User:     "u",
				Password: "p",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := parseDBCredentials(tt.dsn)
			if got := v.GetHost(); got != tt.want.Host {
				t.Errorf("DBOptions.GetHost() = %v, want %v", got, tt.want.Host)
			}
			if got := v.GetPort(); got != tt.want.Port {
				t.Errorf("DBOptions.GetHost() = %v, want %v", got, tt.want.Port)
			}
			if got := v.GetUser(); got != tt.want.User {
				t.Errorf("DBOptions.GetHost() = %v, want %v", got, tt.want.User)
			}
			if got := v.GetPassword(); got != tt.want.Password {
				t.Errorf("DBOptions.GetHost() = %v, want %v", got, tt.want.Password)
			}
			if got := v.GetDBName(); got != tt.want.DBName {
				t.Errorf("DBOptions.GetHost() = %v, want %v", got, tt.want.DBName)
			}
		})
	}
}
