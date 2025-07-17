package memory

import (
	"fmt"
	"log"
	"testing"
)

func TestForReadVal(t *testing.T) {
	log.Println("This goes to stdout")
	pid := 10294                  // Example PID, replace with a valid one for testing
	addr := int64(0x6441c0faa030) // Example address, replace with a valid one for testing
	log.Printf("start to get value")
	got, err := ReadVal[int32](pid, addr)
	if err != nil {
		t.Fatalf("ReadVal failed: %v", err)
	}
	log.Printf("Read value: %d", got)
}
