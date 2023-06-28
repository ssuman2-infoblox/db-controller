package pgbouncer

import (
	"testing"
)

func Test_mockableCmd(t *testing.T) {

	mcmd := &mockableCmd{}
	if _, ok := interface{}(mcmd).(execCmd); !ok {
		t.Error("mockableCmd does not implement execCmd")
	}

}
