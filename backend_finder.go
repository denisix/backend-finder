package main

import (
	"net"
	"strings"
	"time"
	"os"
	"bufio"
	"io"
	"regexp"
	"errors"
	"math/rand"
)

const (
	spawn_delay		= 10 * time.Millisecond // delay between spawning new workers
	timeout_tcp		= 1 * time.Second // timeout on TCP connect
	timeout_read    = 5 * time.Second // timeout reading via TCP socket from server
)

// ----------------------------------------------------------------
func test_server(ret chan<- string, ip, domain, pattern string) bool {

	conn, err := net.DialTimeout("tcp", ip + ":80", timeout_tcp)
	if err == nil { // OK, connected

		str_get := "GET / HTTP/1.0\r\nHost: " + domain + "\r\n\r\n"
		_, err = conn.Write([]byte(str_get))
		if err == nil { // OK, wrote

			reply := make([]byte, 8192)
			conn.SetReadDeadline(time.Now().Add(timeout_read))
			_, err = conn.Read(reply)
			if err == nil {	// OK, replied

				if strings.Contains(string(reply), pattern) { // OK, found pattern

					// println("- found at", ip)
					ret <- "found " + domain + " on IP " + ip
					return true
				}
			}

		}
	}

	ret <- ""
	return false
}

// ----------------------------------------------------------------
func inc(ip net.IP) {
	for j := len(ip)-1; j>=0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ----------------------------------------------------------------
func cidr_to_ips(cidr_raw string) ([]string, error) {
	re := regexp.MustCompile("([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}/[0-9]{1,2})")
	cidr := re.FindStringSubmatch(cidr_raw)

	var ips[]string

	if len(cidr) == 2 {
		ip, ipnet, err := net.ParseCIDR(cidr[1])

		if err != nil {
			println("- Error in CIDR \""+cidr[1]+"\":", err)
			os.Exit(1);
		}
		for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
			ips = append(ips, ip.String())
		}

		// without first .0 and broadcast .255
		return ips[1 : len(ips)-1], nil
	}
	return ips, errors.New("cannot parse CIDR")
}

// ----------------------------------------------------------------
func read_file(fn string) []string {
	f, err := os.Open(fn)
	if err != nil {
		println("- Error opening file \""+ fn + "\":", err)
		os.Exit(1)
	}
	defer f.Close()

	var lines[]string

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString(0x0A) // 0x0A separator = newline
		if err == io.EOF {
			break
		} else if err != nil {
			return nil
		}
		lines = append(lines, line)
	}
	return lines
}

// ----------------------------------------------------------------
func Shuffle(vals []string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for n := len(vals); n > 0; n-- {
		randIndex := r.Intn(n)
		vals[n-1], vals[randIndex] = vals[randIndex], vals[n-1]
	}
}

// ----------------------------------------------------------------
func main() {

	if len(os.Args) != 4 {
		println("Usage: backend_finder.go <domain> <search> <file_CIDRs>")
		os.Exit(1)
	}

	domain := os.Args[1]
	pattern := os.Args[2]
	file_nets := os.Args[3]

	println("Domain:", domain)
	println("Search:", pattern)

	var ips []string
	var cidrs[]string

	cidrs = read_file(file_nets)
	for _,cidr := range cidrs {
		println("- adding CIDR ", cidr)
		add_ips, err := cidr_to_ips(cidr)
		if err != nil {
			println("- Warn: cannot add ["+cidr+"]", err)
		} else {
			ips = append(ips, add_ips...)
		}
	}

	Shuffle(ips)

	/*
	for _,ip := range ips {
		println("ip", ip)
	}
	*/

    var ret chan string = make(chan string, 10)

    println("\n- spawning childs..")

    for _,ip := range ips {

		go test_server(ret, ip, domain, pattern)
		print(".")
		time.Sleep(spawn_delay)
    }
	println("\n- done, scanning..")

	for i := 0; i<len(ips); i++ {
		msg := <-ret
		if msg == "" {
			print(".")
		} else {
	    	print("\n" + msg + "\n")
		}
	}
    println("\nend.")
}
