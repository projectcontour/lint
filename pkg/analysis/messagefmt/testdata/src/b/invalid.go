package b

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"
)

func invalid() {
	verbose := kingpin.Flag("verbose", "verbose mode.").Short('v').Bool() // want `message starts with lowercase: "verbose mode."`

	deleteCommand := kingpin.Command("delete", "delete an object.") // want `message starts with lowercase: "delete an object."`

	fmt.Println(verbose, deleteCommand)
}
