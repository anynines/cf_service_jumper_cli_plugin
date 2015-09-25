package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/a9hcp/cf_service_jumper_cli_plugin/xtunnel"
)

func ListenAndOutputInfo(hosts []string) error {
	var err error

	tunnels := make([]*xtunnel.XTunnel, 0)
	for _, host := range hosts {
		xt := xtunnel.NewUnencryptedXTunnel(host)
		localListenAddress, err := xt.Listen()
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("Listening on %s", localListenAddress))

		tunnels = append(tunnels, xt)
	}

	for _, tunnel := range tunnels {
		go func(tunnel *xtunnel.XTunnel) {
			err = tunnel.Serve()
			if err != nil {
				fmt.Println(fmt.Sprintf("Error on %s: %s", tunnel.LocalAddress(), err))
			}
		}(tunnel)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	_ = <-c

	for _, tunnel := range tunnels {
		err = tunnel.Shutdown()
		if err != nil {
			fmt.Println("[ERR] Failed to shutdown listen socket", err)
		}
	}

	return nil
}
