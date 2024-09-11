package utils

import (
	"flag"
	"fmt"
	"os"
	"unicode"
)

func CustomFlagUsageMessage() {
	fmt.Println("Usage : main.go --dnsPort 53 --serverRefreshTime 10 --ttlForResponse 5")
	flag.PrintDefaults()
	os.Exit(1)
}

func RemoveControlCharacter(data []byte) string {
	result := ""
	for _, b := range data {
		if unicode.IsPrint(rune(b)) || rune(b) == '\n' || rune(b) == '\r' {
			result += string(b)
		}
	}
	return result
}

func PrintHex(data []byte) {
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
