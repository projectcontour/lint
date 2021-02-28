package b

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

func valid() {
	verbose := kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()

	deleteCommand := kingpin.Command("delete", "Delete an object.")

	fmt.Println(verbose, deleteCommand)
}
