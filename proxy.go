/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/apache/openserverless-runtimes/openwhisk"
)

// flag to show version
var version = flag.Bool("version", false, "show version")

// flag to enable debug
var debug = flag.Bool("debug", false, "enable debug output")

// flag to require on-the-fly compilation
var compile = flag.String("compile", "", "compile, reading in standard input the specified function, and producing the result in stdout")

// flag to pass an environment as a json string
var env = flag.String("env", "", "pass an environment as a json string")

// fatal if error
func fatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	// show version number
	if *version {
		fmt.Printf("OpenWhisk ActionLoop Proxy v%s, built with %s\n", openwhisk.Version, runtime.Version())
		return
	}

	// debugging
	if *debug {
		// set debugging flag, propagated to the actions
		openwhisk.Debugging = true
		os.Setenv("OW_DEBUG", "1")
	}

	proxyMode := openwhisk.ProxyModeNone
	useProxyClient := os.Getenv("OW_ACTIVATE_PROXY_CLIENT")
	if useProxyClient == "1" {
		openwhisk.Debug("OW_ACTIVATE_PROXY_CLIENT set. Using runtime as a forward proxy.")
		proxyMode = openwhisk.ProxyModeClient
	}

	useProxyServer := os.Getenv("OW_ACTIVATE_PROXY_SERVER")
	if useProxyServer == "1" {
		openwhisk.Debug("OW_ACTIVATE_PROXY_SERVER set. Using runtime as a proxy server.")
		proxyMode = openwhisk.ProxyModeServer
	}

	// create the action proxy
	ap := openwhisk.NewActionProxy("./action", os.Getenv("OW_COMPILER"), os.Stdout, os.Stderr, proxyMode)

	// compile on the fly upon request
	if *compile != "" {
		ap.ExtractAndCompileIO(os.Stdin, os.Stdout, *compile, *env)
		return
	}

	go hookExitSignals()

	// start the balls rolling
	openwhisk.Debug("OpenWhisk ActionLoop Proxy %s: starting", openwhisk.Version)
	ap.Start(8080)
}

func hookExitSignals() {
	var captureSignalChan = make(chan os.Signal, 1)
	// Relay stop signals to captureSignalChan
	signal.Notify(captureSignalChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGQUIT,
		syscall.SIGHUP)
	signalHandler(<-captureSignalChan)
}

func signalHandler(signal os.Signal) {
	fmt.Printf("\nCaught signal: %+v", signal)
	fmt.Println("\nWait for 1 second to finish processing")
	time.Sleep(1 * time.Second)

	switch signal {

	case syscall.SIGHUP: // kill -SIGHUP XXXX
		fmt.Println("- got hungup")

	case syscall.SIGINT: // kill -SIGINT XXXX or Ctrl+c
		fmt.Println("- got ctrl+c")

	case syscall.SIGTERM: // kill -SIGTERM XXXX
		fmt.Println("- got force stop")

	case syscall.SIGQUIT: // kill -SIGQUIT XXXX
		fmt.Println("- stop and core dump")

	case syscall.SIGABRT: // kill -SIGABRT XXXX
		fmt.Println("- got abort signal")

	case syscall.SIGKILL: // kill -SIGKILL XXXX
		fmt.Println("- got kill signal")

	default:
		fmt.Println("- unknown signal")
	}

	fmt.Println("\nFinished server cleanup")
	os.Exit(0)
}
