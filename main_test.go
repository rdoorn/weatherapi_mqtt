package main

import (
	"log"
	"testing"
)

func TestTimeToEpock(t *testing.T) {
	s := "08:17 AM"
	result := TimeToEpoch(&s)
	log.Printf("%v", result)
}
