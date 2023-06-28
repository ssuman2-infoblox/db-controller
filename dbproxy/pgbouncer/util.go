package pgbouncer

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func parseHostPort(addr string) (host string, port int, err error) {
	parts := strings.SplitN(addr, ":", 2)
	if len(parts) != 2 {
		return "", 0, errors.New("invalid address")
	}
	if ip := net.ParseIP(parts[0]); ip == nil {
		return "", 0, fmt.Errorf("invalid address: %s", parts[0])
	} else {
		host = ip.String()
	}
	if p, err := strconv.ParseUint(parts[1], 10, 16); err != nil {
		return "", 0, fmt.Errorf("invalid port: %s", parts[1])
	} else {
		port = int(p)
	}
	if port <= 0 || port > 65535 {
		return "", 0, fmt.Errorf("invalid port: %d", port)
	}
	return host, port, nil
}

func atomicWrite(filename, contents string) error {
	// Create a temporary file in the same directory as the target file
	tempFile, err := ioutil.TempFile(filepath.Dir(filename), "temp")
	if err != nil {
		return err
	}

	// Write the contents to the temporary file
	_, err = tempFile.WriteString(contents)
	if err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return err
	}

	// Close the temporary file to flush the contents to disk
	err = tempFile.Close()
	if err != nil {
		os.Remove(tempFile.Name())
		return err
	}

	// Atomically rename the temporary file to the target filename
	err = os.Rename(tempFile.Name(), filename)
	if err != nil {
		os.Remove(tempFile.Name())
		return err
	}

	return nil
}
