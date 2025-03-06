package sshserver

import (
	"fmt"
	"testing"
)

func TestFindSetVfwInfo(t *testing.T) {

	vfwinfo, err := FindSetVfwInfo("aaatest")
	if err != nil {
		t.Fatalf("Unexpected error; %v", err)
	}
	fmt.Println(vfwinfo.IP)
}
