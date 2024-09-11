package ipProcessing

import (
	"errors"
	"fmt"
	"github.com/hkiiita/GopherDNS/config"
	"net"
)

func GetIpAddresses(domainQuery string) (string, error) {
	for domain, ipAddress := range config.GetDomainConfig().DomainIpMapping {
		if domain == domainQuery {
			return ipAddress, nil
		}
	}
	return "", errors.New("domain not found")
}

func GetIP4ByteArray(ipStr string) ([4]byte, error) {
	// Parse the IP address from the string
	ip := net.ParseIP(ipStr)

	if ip == nil {
		fmt.Println("Invalid IP address")
		return [4]byte{}, errors.New("Invalid IP address")
	}

	//IP address to a byte slice
	ipv4 := ip.To4()

	//if the IP address is IPv4
	if ipv4 == nil {
		fmt.Println("Not an IPv4 address")
		return [4]byte{}, errors.New("Not an IPv4 address")
	}

	//byte slice to [4]byte array
	var ipArray [4]byte
	copy(ipArray[:], ipv4)

	return ipArray, nil
}
