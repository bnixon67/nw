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
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
)

// checkClose will attempt to Close the given resource checking for any error
// although it is unlikely there will be an error with Close, it doesn't hurt to check
func checkClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// receiveFile listens on the host and port for a file.
func receiveFile(host, port, fileName string, overwrite bool) (written int64, err error) {
	var addr = net.JoinHostPort(host, port)

	log.SetPrefix(fmt.Sprintf("Recv(%d) ", os.Getpid()))
	log.Println("Listening on", addr)

	// start server
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer checkClose(ln)

	// accept connection
	conn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer checkClose(conn)

	// create file or use stdout if no file specified
	var file *os.File
	if fileName == "" {
		file = os.Stdout
	} else {
		// check if file exists
		_, err = os.Stat(fileName)
		if err == nil {
			// file exists
			if !overwrite {
				log.Printf("File %s exists and overwrite not set", fileName)
				return 0, os.ErrExist
			}
		} else if os.IsNotExist(err) {
			// file does not exist, so fall through
		} else {
			// undefined error
			log.Println("undefined error:", err)
			return 0, err
		}

		file, err = os.Create(fileName)
		if err != nil {
			log.Println("os.Create failed:", err)
			return 0, err
		}
	}

	log.Println("Receive writing to", file.Name())
	// copy everythin from connection to file
	written, err = io.Copy(file, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received %d bytes", written)

	return written, err
}

// sendFile send a file via the host and port.
func sendFile(host, port, fileName string) (sent int64, err error) {
	var addr = host + ":" + port

	log.SetPrefix(fmt.Sprintf("Send(%d) ", os.Getpid()))

	// dial server
	log.Println("Sending to", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer checkClose(conn)

	// open file or use stdin if no file specified
	var file *os.File
	if fileName == "" {
		file = os.Stdin
	} else {
		file, err = os.Open(fileName)
		if err != nil {
			log.Println(err)
			return 0, err
		}
		defer checkClose(file)
	}

	// copy file to connection
	log.Println("Sending file", file.Name())
	written, err := io.Copy(conn, file)
	if err != nil {
		log.Println(err)
		return written, err
	}
	log.Printf("Sent %d bytes", written)

	return written, err
}

func main() {
	// define command-line flags
	var (
		host        string
		port        string
		overwrite   bool
		logFileName string
		logFileMode int
	)
	flag.StringVar(&host, "host", "localhost", "the host to send to or receive from")
	flag.StringVar(&port, "port", "4200", "the port to send to or receive from")
	flag.BoolVar(&overwrite, "overwrite", false, "overwrite existing file")
	flag.StringVar(&logFileName, "logFileName", "", "the file name of the log file")
	flag.IntVar(&logFileMode, "logFileMode", 0600, "the FileMode for the log file")

	// customize Usage function for command-line parsing
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s [flags] send|receive\n", os.Args[0])
		flag.PrintDefaults()
	}

	// parse commandline flags
	flag.Parse()

	// setup logging
	//log.SetFlags(log.LstdFlags | log.Lmsgprefix | log.Lshortfile)
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)

	// log to a file, if given
	if logFileName != "" {
		// append or create file for read-write
		logFile, err := os.OpenFile(logFileName,
			os.O_RDWR|os.O_CREATE|os.O_APPEND,
			fs.FileMode(logFileMode))
		if err != nil {
			log.Fatal(err)
		}
		defer checkClose(logFile)

		log.SetOutput(logFile)
	}

	// determine mode to run, i.e. receive or send
	switch flag.Arg(0) {
	case "receive":
		_, err := receiveFile(host, port, flag.Arg(1), overwrite)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Receive failed", err)
			os.Exit(3)
		}

	case "send":
		sendFile(host, port, flag.Arg(1))

	case "":
		flag.Usage()
		fmt.Fprintln(flag.CommandLine.Output(),
			"Need to provide either send or receive")
		os.Exit(1)

	default:
		flag.Usage()
		fmt.Fprintln(flag.CommandLine.Output(),
			"Invalid command, provide either send or receive")
		os.Exit(2)
	}

}
