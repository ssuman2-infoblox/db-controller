package pgbouncer

import "testing"

func Test_parseHostPort(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantHost string
		wantPort int
	}{
		{
			name:    "empty",
			args:    args{addr: ""},
			wantErr: true,
		},
		{
			name:    "0 port",
			args:    args{addr: "127.0.0.1:0"},
			wantErr: true,
		},
		{
			name:     "80 port",
			args:     args{addr: "127.0.0.1:80"},
			wantErr:  false,
			wantHost: "127.0.0.1",
			wantPort: 80,
		},
		{
			name:    "hostname",
			args:    args{addr: "localhost:80"},
			wantErr: true,
		},
		{
			name:    "badport",
			args:    args{addr: "127.0.0.1:f"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort, err := parseHostPort(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHostPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHost != tt.wantHost {
				t.Errorf("parseHostPort() gotHost = %v, want %v", gotHost, tt.wantHost)
				return
			}
			if gotPort != tt.wantPort {
				t.Errorf("parseHostPort() gotPort = %v, want %v", gotPort, tt.wantPort)
				return
			}
		})
	}
}
