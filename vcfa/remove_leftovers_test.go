package vcfa

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

// This file contains routines that clean up the test suite after failed tests

func removeLeftovers(govcdClient *govcd.VCDClient, verbose bool) error {
	if verbose {
		fmt.Printf("Start leftovers removal\n")
	}

	return nil
}
