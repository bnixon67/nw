/*
Copyright 2021 Bill Nixon

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

// These are the default testing constants.
const (
	host      = "localhost"
	port      = "4200"
	fileName  = ""
	overwrite = false
)

// compareFile compares two files.
// Assumes the files are small since the comapre is done in memory.
func compareFile(fileName1, fileName2 string) (equal bool, err error) {
	file1, err := os.ReadFile(fileName1)
	if err != nil {
		return false, err
	}

	file2, err := os.ReadFile(fileName2)
	if err != nil {
		return false, err
	}

	return bytes.Equal(file1, file2), nil
}

// TestMain to set log flags and log output file.
func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix | log.Lshortfile)

	// append or create file for read-write
	logFile, err := os.OpenFile("nw_test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer checkClose(logFile)
	log.SetOutput(logFile)

	os.Exit(m.Run())
}

func TestReceiveInvalidHost(t *testing.T) {
	_, err := receiveFile("nosuchhost", port, fileName, overwrite)
	if err == nil {
		t.Errorf("Expected err, got <nil>")
	}
}

func TestReceiveInvalidPort(t *testing.T) {
	_, err := receiveFile(host, "a", fileName, overwrite)
	if err == nil {
		t.Errorf("Expected err, got <nil>")
	}
}

func TestReceiveInvalidPath(t *testing.T) {
	// run receiveFile in background via go routine
	// use WaitGroup to wait until receiveFile finishes
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, err := receiveFile(host, port, "/foo/bar", true)
		if err == nil {
			t.Errorf("Expected err, got <nil>")
		}
		wg.Done()
	}()

	// allow time to startup receive server
	time.Sleep(500 * time.Millisecond)

	sendFile(host, port, "nw.go")

	// wait for receiveFile
	wg.Wait()
}

func TestReceiveOverwrite(t *testing.T) {
	// create temp file
	f, err := os.CreateTemp("", "nw_test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// run receiveFile in background via go routine
	// use WaitGroup to wait until receiveFile finishes
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, err := receiveFile(host, port, f.Name(), false)
		if err == nil {
			t.Errorf("Expected err, got <nil>")
		}
		wg.Done()
	}()

	// allow time for receiveFile to start
	time.Sleep(500 * time.Millisecond)

	sendFile(host, port, "nw.go")

	// wait for receiveFile
	wg.Wait()
}

func TestSendFile(t *testing.T) {
	// create temp file
	f, err := os.CreateTemp("", "nw_test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	var received int64
	var receiveErr error

	sendFileName := "nw.go"
	recvFileName := f.Name()

	// run receiveFile in background via go routine
	// use WaitGroup to wait until receiveFile finishes
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		received, receiveErr = receiveFile(host, port, recvFileName, true)
		if receiveErr != nil {
			t.Error("Expected err == <nil>, got", err)
		}
		wg.Done()
	}()

	// allow time for receiveFile to start
	time.Sleep(500 * time.Millisecond)

	sent, sentErr := sendFile(host, port, sendFileName)
	if sentErr != nil {
		t.Error("Expected err == <nil>, got", sentErr)
	}

	// wait for receiveFile
	wg.Wait()

	if sent != received {
		t.Errorf("Expected sent = received, got %d = %d\n", sent, received)
	}

	equal, err := compareFile(sendFileName, recvFileName)
	if err != nil {
		t.Error("compareFile failed")
	}
	if !equal {
		t.Error("files did not match")
	}
}

func TestSendNoReceiver(t *testing.T) {
	_, err := sendFile(host, port, fileName)
	if err == nil {
		t.Errorf("Expected err, got <nil>")
	}
}
