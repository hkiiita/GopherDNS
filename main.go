package main

import (
	"errors"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"net"
	"os"
	"time"
	"unicode"
)

type DomainConfig struct {
	DomainIpMapping map[string]string `yaml:"domains"`
}

var domainValues *DomainConfig

func customFlagUsageMessage() {
	fmt.Println("Usage : main.go --dnsPort 53 --serverRefreshTime 10 --ttlForResponse 5")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var dnsPort = flag.Int("dnsPort", 0, "DNS port.")
	var serverRefreshTime = flag.Int("serverRefreshTime", 0, "Time under which server would periodically refresh domain IPs")
	var ttlForResponse = flag.Int("ttlForResponse", 0, "time to live to be sent in DNS response")

	flag.Usage = customFlagUsageMessage
	flag.Parse()

	if *dnsPort == 0 || *serverRefreshTime == 0 || *ttlForResponse == 0 {
		flag.Usage()
	}

	var err error
	domainValues, err = readConfig()
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
				domainValues, err = readConfig()
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

func processRequest(request []byte, conn *net.UDPConn, address *net.UDPAddr, ttlForResponse int) {
	log.Printf("Received request... processing...")
	domainQuery, err := extractDomainName(request)
	if err != nil {
		log.Printf("Domain not found. %+v", err)
		return
	}

	ip, err := getIpAddresses(domainQuery)
	if err != nil {
		log.Printf("IP not found. %+v", err)
		return
	}

	resp, err := createResponse(request, ip, ttlForResponse)
	if err != nil {
		log.Printf("Unable to create response. %+v", err)
		return
	}

	printHex(resp)

	sendResponse(resp, address, conn)
}

func printHex(data []byte) {
	for i, b := range data {
		if i%16 == 0 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("%04x: ", i)
		}
		fmt.Printf("%02x ", b)
	}
	fmt.Println()
}

func sendResponse(response []byte, address *net.UDPAddr, conn *net.UDPConn) {
	_, err := conn.WriteToUDP(response, address)
	if err != nil {
		log.Printf("Error sending DNS response : %+v", err)
	}
}

func createResponse(query []byte, ipAddress string, ttlForResponse int) ([]byte, error) {
	var response []byte
	//header = append(header, query[0:1]...) // Added transaction ID

	// Copy the query ID
	response = append(response, query[0:2]...)

	// Set response flags: 0x8180 (Standard query response, No error)
	response = append(response, 0x81, 0x80)

	// Set QDCOUNT (number of questions), ANCOUNT (number of answers)
	response = append(response, 0x00, 0x01) // QDCOUNT
	response = append(response, 0x00, 0x01) // ANCOUNT

	// No NSCOUNT, ARCOUNT
	response = append(response, 0x00, 0x00) // NSCOUNT
	response = append(response, 0x00, 0x00) // ARCOUNT

	// Copy the query section
	response = append(response, query[12:]...)

	// Answer section
	// Name
	response = append(response, 0xc0, 0x0c) // Name (same as question)
	// Type A (0x0001)
	response = append(response, 0x00, 0x01)
	// Class IN (0x0001)
	response = append(response, 0x00, 0x01)
	// TTL (0x00000000)
	response = append(response, 0x00, 0x00, 0x00, 0x40)
	// Length of IP address (4 bytes)
	response = append(response, 0x00, 0x04)
	// IP address
	ip := net.ParseIP(ipAddress).To4()
	if ip == nil {
		return nil, errors.New("Invaid IP")
	}
	response = append(response, ip...)
	return response, nil
}

// TODO: This can be done more cleanly as a function returning next offset to be followed and number of bytes to be read.
func extractDomainName(dnsMessage []byte) (string, error) {
	if len(dnsMessage) < 12 {
		log.Printf("Error | DNS message is too short.")
		return "", errors.New("DNS message too short")
	}

	// The header of DNS message is standard 12 bytes , as of now we do not need header.
	// but would need to parse it to know the value of `Questions` filed to get number of domains being queried.
	// Hence, here we just extract the remaining stuff.
	dnsMessageQueriesBytes := dnsMessage[12:]

	// After the header ends, the Query section begins.
	numberOfSubDomainBytes := dnsMessageQueriesBytes[0]
	if numberOfSubDomainBytes == 0 {

	}
	subDomainBytes := dnsMessageQueriesBytes[1 : numberOfSubDomainBytes+1]
	log.Printf("SUB DOMAIN : %+v", string(subDomainBytes))

	numberOfDomainBytes := dnsMessageQueriesBytes[numberOfSubDomainBytes+1]
	if numberOfDomainBytes == 0 {

	}
	domainBytesBeginIndex := numberOfSubDomainBytes + 2
	domainBytes := dnsMessageQueriesBytes[domainBytesBeginIndex : domainBytesBeginIndex+numberOfDomainBytes+1]
	log.Printf("DOMAIN : %+v", string(domainBytes))

	numberOfTopLevelDomainBytes := dnsMessageQueriesBytes[numberOfSubDomainBytes+1+numberOfDomainBytes+1]
	if numberOfTopLevelDomainBytes == 0 {

	}
	topLevelDomainBeginIndex := numberOfSubDomainBytes + 1 + numberOfDomainBytes + 2
	topLevelDomainBytes := dnsMessageQueriesBytes[topLevelDomainBeginIndex : topLevelDomainBeginIndex+numberOfTopLevelDomainBytes+1]
	log.Printf("Top Level Domain : %+v ", string(topLevelDomainBytes))

	return removeControlCharacter(domainBytes), nil
}

func readConfig() (*DomainConfig, error) {
	// Load the YAML configuration
	file, err := os.Open("domains.yaml")
	if err != nil {
		log.Fatalf("Error opening YAML file: %v", err)
		return nil, err
	}
	defer file.Close()

	var config DomainConfig
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding YAML file: %v", err)
		return nil, err
	}
	return &config, nil
}

func getIpAddresses(domainQuery string) (string, error) {
	for domain, ipAddress := range domainValues.DomainIpMapping {
		if domain == domainQuery {
			return ipAddress, nil
		}
	}
	return "", errors.New("domain not found")
}

func removeControlCharacter(data []byte) string {
	result := ""
	for _, b := range data {
		if unicode.IsPrint(rune(b)) || rune(b) == '\n' || rune(b) == '\r' {
			result += string(b)
		}
	}
	return result
}
