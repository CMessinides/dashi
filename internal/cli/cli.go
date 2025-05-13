package cli

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/cmessinides/dashi/internal/server"
)

type Config struct {
	Dev bool
}

func Run(conf Config) int {
	f := flag.NewFlagSet("dashi", flag.ContinueOnError)
	addr := f.String("addr", ":8080", "address to listen on")

	if err := f.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		} else {
			fmt.Println(err)
			return 2
		}
	}

	s := server.NewServer(server.Config{
		Dev: conf.Dev,
	})

	fmt.Printf("Dashi is listening on %s\n", *addr)
	err := http.ListenAndServe(*addr, s)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}
