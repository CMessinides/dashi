package cli

import (
	"errors"
	"flag"
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
			return 2
		}
	}

	s := server.NewServer(server.Config{
		Dev: conf.Dev,
	})

	if err := s.Run(*addr); err != nil {
		return 1
	}

	return 0
}
