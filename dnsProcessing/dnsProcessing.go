package dnsProcessing

import (
	"errors"
	"github.com/hkiiita/GopherDNS/ipProcessing"
	"github.com/hkiiita/GopherDNS/utils"
	"golang.org/x/net/dns/dnsmessage"
	"log"
	"net"
)

func SendResponse(response []byte, address *net.UDPAddr, conn *net.UDPConn) {
	_, err := conn.WriteToUDP(response, address)
	if err != nil {
		log.Printf("Error sending DNS response : %+v", err)
	}
}

func CreateResponseMessage(parsedHeader dnsmessage.Header, parsedQuestion dnsmessage.Question, ttlForResponse int) ([]byte, error) {
	ipStr, err := ipProcessing.GetIpAddresses(parsedQuestion.Name.String()[:len(parsedQuestion.Name.String())-1])
	if err != nil {
		log.Printf("Error IP not found. %+v", err)
		return nil, err
	}

	ip4Bytes, err := ipProcessing.GetIP4ByteArray(ipStr)
	if err != nil {
		log.Printf("Error converting IP to 4 byte slice. %+v", err)
		return nil, err
	}

	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:                 parsedHeader.ID,
			Response:           true,
			OpCode:             0,
			Authoritative:      true,
			Truncated:          false,
			RecursionDesired:   false,
			RecursionAvailable: false,
			AuthenticData:      false,
			CheckingDisabled:   false,
			RCode:              0,
		},
		Questions: []dnsmessage.Question{
			{
				Name:  parsedQuestion.Name,
				Type:  parsedQuestion.Type,
				Class: parsedQuestion.Class,
			},
		},
		Answers: []dnsmessage.Resource{
			{
				Header: dnsmessage.ResourceHeader{
					Name:  parsedQuestion.Name,
					Type:  dnsmessage.TypeA,
					Class: dnsmessage.ClassINET,
					TTL:   uint32(ttlForResponse),
				},
				Body: &dnsmessage.AResource{
					A: ip4Bytes,
				},
			},
		},
		Authorities: make([]dnsmessage.Resource, 0),
		Additionals: make([]dnsmessage.Resource, 0),
	}

	buf, err := msg.Pack()
	if err != nil {
		panic(err)
	}

	return buf, nil
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

	return utils.RemoveControlCharacter(domainBytes), nil
}
