package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/oleksandr/bonjour"
	"github.com/pkg/errors"
)

var timeout int

func init() {
	flag.IntVar(&timeout, "t", 5, "Number of seconds to timeout.")
}

func main() {
	code, err := run()
	if err != nil {
		fmt.Printf("Error : %v\n", err)
	}
	os.Exit(code)
}

func run() (int, error) {
	flag.Parse()
	if timeout < 0 {
		timeout = 0
	}

	resolver, err := bonjour.NewResolver(nil)
	if err != nil {
		return 1, errors.Wrap(err, "Error NewResolver.")
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	results := make(chan *bonjour.ServiceEntry)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		fmt.Println("Interface            HostName             IPAddress(v4)  IPAddress(v6) ")
	LOOP:
		for {
			select {
			case e := <-results:
				var ipv4, ipv6 string
				if e.AddrIPv4 != nil {
					ipv4 = e.AddrIPv4.String()
				}
				if e.AddrIPv6 != nil {
					ipv6 = e.AddrIPv6.String()
				}
				fmt.Printf("%-20s %-20s %-14s %-14s\n", e.Instance, e.HostName, ipv4, ipv6)
			case <-sigCh:
				break LOOP
			case <-time.After(time.Duration(timeout) * time.Second):
				break LOOP
			}
		}
		resolver.Exit <- true
		wg.Done()
	}()

	err = resolver.Browse("_irkit._tcp", "local.", results)
	if err != nil {
		return 1, errors.Wrap(err, "Failed to brows.")
	}

	wg.Wait()

	return 0, nil
}
