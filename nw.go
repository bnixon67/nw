package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

func start_receive(host, port, tempFileDir, tempFilePattern string) {
	var addr = host + ":" + port

	log.Println("Listen", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	conn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connection accepted")

	tmpFile, err := ioutil.TempFile(tempFileDir, tempFilePattern)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("TempFile", tmpFile.Name())

	written, err := io.Copy(tmpFile, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connection done, bytes received", written)

	conn.Close()
}

func start_send(host, port, fileName string) {
	var addr = host + ":" + port

	log.Println("Dial", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Println("Start copy of", fileName)
	written, err := io.Copy(conn, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("End copy, bytes written", written)
}

func main() {
	var host, port, tempFileDir, tempFilePattern string

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s [flags] send|receive\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&host, "host", "localhost", "the host to send to or receive from")
	flag.StringVar(&port, "port", "4200", "the port to send to or receive from")
	flag.StringVar(&tempFileDir, "tmpDir", os.TempDir(), "directory for temportary files")
	flag.StringVar(&tempFilePattern, "tmpPattern", "nw.", "pattern for temportary files")

	flag.Parse()

	/*
		if flag.NArg() != 1 {
			flag.Usage()
			os.Exit(1)
		}
	*/

	switch flag.Arg(0) {
	case "receive":
		start_receive(host, port, tempFileDir, tempFilePattern)

	case "send":
		start_send(host, port, flag.Arg(1))

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
