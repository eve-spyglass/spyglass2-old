package nats

import (
	"flag"
	"fmt"
	natsd "github.com/nats-io/nats-server/v2/server"
	"os"
	"strconv"
)

type (
	MessageServer struct {
		server *natsd.Server
	}
)

func New(port int) (*MessageServer, error) {

	args := []string{"nats-server", "-a", "127.0.0.1", "-p", strconv.Itoa(port)}

	fs := flag.NewFlagSet("nats-server", flag.ExitOnError)
	fs.Usage = func() {
		os.Exit(0)
	}
	opts, err := natsd.ConfigureOptions(fs, args[1:], natsd.PrintServerAndExit, fs.Usage, natsd.PrintTLSHelpAndDie)
	if err != nil{
		return nil, fmt.Errorf("failed to configure nats-server: %w", err)
	}

	if err != nil {
		natsd.PrintAndDie(fmt.Sprintf("%s: %s", "nats", err))
	}


	srv, err := natsd.NewServer(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create nats server: %w", err)
	}
	srv.ConfigureLogger()

	msgsrv := &MessageServer{
		server: srv,
	}

	return msgsrv, nil
}

func (srv *MessageServer)Run() error {
	if err := natsd.Run(srv.server); err != nil {
		return fmt.Errorf("nats-server error: %w" ,err)
	}
	srv.server.WaitForShutdown()
	return nil
}
