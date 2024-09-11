package main

import (
	"flag"
	"github.com/hkiiita/GopherDNS/config"
	"github.com/hkiiita/GopherDNS/dnsProcessing"
	"github.com/hkiiita/GopherDNS/utils"
	"golang.org/x/net/dns/dnsmessage"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	var dnsPort = flag.Int("dnsPort", 0, "DNS port.")
	var serverRefreshTime = flag.Int("serverRefreshTime", 0, "Time under which server would periodically refresh domain IPs")
	var ttlForResponse = flag.Int("ttlForResponse", 0, "time to live to be sent in DNS response")

	flag.Usage = utils.CustomFlagUsageMessage
	flag.Parse()

	if *dnsPort == 0 || *serverRefreshTime == 0 || *ttlForResponse == 0 {
		flag.Usage()
	}

	var err error
	err = config.SetDomainConfig()
	if err != nil {
		log.Fatalf("Error getting IP values for domains %+v", err)
	}

	ticker := time.NewTicker(time.Duration(*serverRefreshTime) * time.Second)
	defer func() {
		ticker.Stop()
	}()

	go func() {
		for {
			select {
			case <-ticker.C:
				err = config.SetDomainConfig()
				if err != nil {
					log.Fatalf("Error getting IP values for domains %+v", err)
				}
			}
		}
	}()

	// UDP end point.
	addr := net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: *dnsPort,
	}

	// open up a process listening on above address.
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("Error creating a UDP connection %+v", err)
		os.Exit(1)
	}
	defer conn.Close()

	log.Printf("DNS server listening on port %+v", addr.Port)

	// Read requests
	for {
		buf := make([]byte, 512)
		log.Printf("waiting for requests...")
		n, address, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading incoming reuqest into buffer : %+v", err)
			continue
		}
		go processRequest(buf[:n], conn, address, *ttlForResponse)
	}
}

func processRequest(dnsRequest []byte, conn *net.UDPConn, address *net.UDPAddr, ttlForResponse int) {
	log.Printf("Received request... processing...")

	parser := dnsmessage.Parser{}
	parsedHeader, err := parser.Start(dnsRequest)
	parsedQuestion, err := parser.Question()

	resp, err := dnsProcessing.CreateResponseMessage(parsedHeader, parsedQuestion, ttlForResponse)
	if err != nil {
		log.Printf("Unable to create response. %+v", err)
		return
	}

	utils.PrintHex(resp)

	dnsProcessing.SendResponse(resp, address, conn)
}
