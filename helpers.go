package main

import (
	"encoding/hex"
	"math/rand"
	"regexp"
)

// thank you from http://stackoverflow.com/questions/45267125/ddg#59457748
func generateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func testMatch(input, regex string) bool {
	match, _ := regexp.MatchString(
		regex,
		input,
	) // we can ignore the error since we know that the regex is valid
	return match
}
