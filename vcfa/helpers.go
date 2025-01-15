package vcfa

import (
	"github.com/vmware/go-vcloud-director/v3/util"
	"os"
)

// safeClose closes a file and logs the error, if any. This can be used instead of file.Close()
func safeClose(file *os.File) {
	if err := file.Close(); err != nil {
		util.Logger.Printf("Error closing file: %s\n", err)
	}
}
