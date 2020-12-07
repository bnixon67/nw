package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// do_receive listens on the host and port for a file.
func do_receive(host, port, fileName string, overwrite bool) {
	var addr = host + ":" + port

	// start server
	log.Println("Listen", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	// accept connection
	conn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connection accepted")
	defer conn.Close()

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
				log.Fatal("file exists and overwrite not set")
			}
		} else if os.IsNotExist(err) {
			// file does not exist, so fall through
		} else {
			// undefined error
			log.Fatal(err)
		}

		file, err = os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("File", file.Name())

	// copy everythin from connection to file
	written, err := io.Copy(file, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connection done, bytes received", written)
}

// do_send send a file via the host and port.
func do_send(host, port, fileName string) {
	var addr = host + ":" + port

	// dial server
	log.Println("Dial", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// open file or use stdin if no file specified
	var file *os.File
	if fileName == "" {
		file = os.Stdin
	} else {
		file, err = os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}

	// copy file to connection
	log.Println("Start copy of", file.Name())
	written, err := io.Copy(conn, file)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("End copy, bytes written", written)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// customize Usage function for command-line parsing
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s [flags] send|receive\n", os.Args[0])
		flag.PrintDefaults()
	}

	// define command-line flags
	var (
		host      string
		port      string
		overwrite bool
	)
	flag.StringVar(&host, "host", "localhost", "the host to send to or receive from")
	flag.StringVar(&port, "port", "4200", "the port to send to or receive from")
	flag.BoolVar(&overwrite, "overwrite", false, "overwrite existing file")

	// parse commandline flags
	flag.Parse()

	// determine mode to run, i.e. receive or send
	switch flag.Arg(0) {
	case "receive":
		do_receive(host, port, flag.Arg(1), overwrite)

	case "send":
		do_send(host, port, flag.Arg(1))

	case "":
		flag.Usage()
		fmt.Fprintln(flag.CommandLine.Output(), "Need to provide either send or receive")
		os.Exit(1)

	default:
		flag.Usage()
		fmt.Fprintln(flag.CommandLine.Output(), "Invalid command, provide either send or receive")
		os.Exit(2)
	}

}
