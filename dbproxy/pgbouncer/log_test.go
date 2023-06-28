package pgbouncer

import (
	"reflect"
	"testing"
)

type logLine struct {
	level string
	line  string
}

type testLogReceiver struct {
	lines []logLine
}

func (t *testLogReceiver) Debug(args ...interface{}) {
	t.lines = append(t.lines, logLine{"DEBUG", args[0].(string)})
}

func (t *testLogReceiver) Info(args ...interface{}) {
	t.lines = append(t.lines, logLine{"INFO", args[0].(string)})
}

func (t *testLogReceiver) Warn(args ...interface{}) {
	t.lines = append(t.lines, logLine{"WARN", args[0].(string)})
}

func (t *testLogReceiver) Error(args ...interface{}) {
	t.lines = append(t.lines, logLine{"ERROR", args[0].(string)})
}

func (t *testLogReceiver) Fatal(args ...interface{}) {
	t.lines = append(t.lines, logLine{"FATAL", args[0].(string)})
}

func Test_pgbouncerLogger_Write(t *testing.T) {

	tests := []struct {
		name      string
		line      []byte
		wantErr   bool
		wantLines []logLine
	}{
		{
			name:    "empty",
			line:    []byte(""),
			wantErr: false,
		},
		{
			name:    "error line",
			line:    []byte(`2023-07-01 11:41:56.659 CDT [60355] ERROR this is an error`),
			wantErr: false,
			wantLines: []logLine{
				{"ERROR", `this is an error`},
			},
		},
		{
			name:    "info line",
			line:    []byte(`2023-07-01 11:41:56.659 CDT [60355] LOG this an info line`),
			wantErr: false,
			wantLines: []logLine{
				{"INFO", `this an info line`},
			},
		},
		{
			name:    "info line 2",
			line:    []byte(`2023-07-01 11:41:56.659 CDT [60355] INFO this an info line`),
			wantErr: false,
			wantLines: []logLine{
				{"INFO", `this an info line`},
			},
		},
		{
			name:    "fatal line",
			line:    []byte(`2023-07-01 11:56:17.426 CDT [61406] FATAL cannot load config file`),
			wantErr: false,
			wantLines: []logLine{
				{"ERROR", `cannot load config file`},
				{"ERROR", `pgbouncer exited with FATAL error`},
			},
		},
		{
			name:    "info with newline",
			line:    []byte("2023-07-01 11:41:56.659 CDT [60355] INFO this an info line\n"),
			wantErr: false,
			wantLines: []logLine{
				{"INFO", `this an info line`},
			},
		},
		{
			name:    "debug line",
			line:    []byte(`2023-07-01 11:41:56.659 CDT [60355] DEBUG this an debug line`),
			wantErr: false,
			wantLines: []logLine{
				{"DEBUG", `this an debug line`},
			},
		},
		{
			name:    "warning line",
			line:    []byte(`2023-07-01 11:41:56.659 CDT [60355] WARNING this an warn line`),
			wantErr: false,
			wantLines: []logLine{
				{"WARN", `this an warn line`},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := &testLogReceiver{}
			l := &pgbouncerLogger{
				log: receiver,
			}
			gotN, err := l.Write(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("pgbouncerLogger.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			wantN := len(tt.line)
			if gotN != wantN {
				t.Errorf("pgbouncerLogger.Write() = %v, want %v", gotN, wantN)
			}
			if !reflect.DeepEqual(receiver.lines, tt.wantLines) {
				t.Errorf("pgbouncerLogger.Write() lines = %v, want %v", receiver.lines, tt.wantLines)
			}
		})
	}
}
