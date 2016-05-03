package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anynines/cf_service_jumper_cli_plugin/xtunnel"
)

func ListenAndOutputInfo(hosts []string, sharedSecret string, connectionPrinter ConnectionPrinter) error {
	var err error

	identity, key, err := GetIdentityAndKey(sharedSecret)
	if err != nil {
		return err
	}

	tunnels := make([]*xtunnel.XTunnel, 0)
	for _, host := range hosts {
		xt := xtunnel.NewXTunnelPSK("localhost:0", host, identity, key)
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

	var sampleOutputs []string
	for _, tunnel := range tunnels {
		sampleOutput := connectionPrinter.SampleCallOutput(tunnel.LocalAddress())
		if len(sampleOutput) > 0 {
			sampleOutputs = append(sampleOutputs, sampleOutput)
		}
	}
	OutputSampleCmds(sampleOutputs)

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
