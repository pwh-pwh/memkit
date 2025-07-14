package pid

import (
	"log"
	"testing"
)

func TestPidofPidFetcher_GetPID(t *testing.T) {
	packageName := "test_cli"
	p := PidofPidFetcher{}
	got, err := p.GetPID(packageName)
	if err != nil {
		t.Fatal(err)
	}
	if got == -1 {
		t.Fatalf("pid of %s is -1", packageName)
	}
	log.Printf("pid of %s is %d", packageName, got)
}

func TestProcPidFetcher_GetPID(t *testing.T) {
	packageName := "test_cli"
	p := PsPidFetcher{}
	got, err := p.GetPID(packageName)
	if err != nil {
		t.Fatal(err)
	}
	if got == -1 {
		t.Fatalf("pid of %s is -1", packageName)
	}
	log.Printf("pid of %s is %d", packageName, got)
}

func TestPsPidFetcher_GetPID(t *testing.T) {
	packageName := "test_cli"
	p := ProcPidFetcher{}
	got, err := p.GetPID(packageName)
	if err != nil {
		t.Fatal(err)
	}
	if got == -1 {
		t.Fatalf("pid of %s is -1", packageName)
	}
	log.Printf("pid of %s is %d", packageName, got)
}
