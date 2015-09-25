package xtunnel

import (
	"fmt"
	"net"

	"github.com/raff/tls-ext"
	"github.com/raff/tls-psk"
)

type XTunnel struct {
	localService  string
	remoteService string
	localListener net.Listener
	config        *tls.Config
}

func NewUnencryptedXTunnel(remoteService string) *XTunnel {
	return createXTunnel("localhost:0", remoteService, nil)
}

// NewXTunnel creates a new XTunnel instance using certificate based TLS
func NewXTunnel(localService, remoteService string) *XTunnel {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	return createXTunnel(localService, remoteService, config)
}

// NewXTunnelPSK creates a new XTunnel instance using TLS-PSK
func NewXTunnelPSK(localService, remoteService, pskIdentity, pskey string) *XTunnel {
	config := &tls.Config{
		CipherSuites: []uint16{psk.TLS_PSK_WITH_AES_128_CBC_SHA, psk.TLS_PSK_WITH_AES_256_CBC_SHA},
		Extra: psk.PSKConfig{
			GetKey: func(id string) ([]byte, error) {
				return []byte(pskey), nil
			},
			GetIdentity: func() string {
				return pskIdentity
			},
		},
	}

	return createXTunnel(localService, remoteService, config)
}

// Listen creates the listening socket.
func (xt *XTunnel) Listen() (string, error) {
	var err error
	xt.localListener, err = net.Listen("tcp", xt.localService)
	if err != nil {
		return "", fmt.Errorf("[ERR] Failed to listen on a random tcp socket. %s", err)
	}
	return xt.localListener.Addr().String(), nil
}

// Serve waits for client connections to be processed. Blocks!
func (xt *XTunnel) Serve() error {
	for {
		// wait until a client connects
		conn, err := xt.localListener.Accept()
		if err != nil {
			return err
		} else {
			// process the clients request
			err = xt.createConnPipe(conn)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Shutdown performs cleanup.
func (xt *XTunnel) Shutdown() error {
	err := xt.localListener.Close()
	return err
}

func createXTunnel(localService, remoteService string, config *tls.Config) *XTunnel {
	return &XTunnel{
		localService:  localService,
		remoteService: remoteService,
		config:        config,
	}
}

func (xt *XTunnel) createConnPipe(localConn net.Conn) error {
	var err error
	var remoteConn net.Conn
	if xt.config == nil {
		remoteConn, err = net.Dial("tcp", xt.remoteService)
	} else {
		remoteConn, err = tls.Dial("tcp", xt.remoteService, xt.config)
	}

	if err != nil {
		return err
	}

	go func() {
		Pipe(localConn, remoteConn)
		// close connections when we're done
		defer localConn.Close()
		defer remoteConn.Close()
	}()

	return nil
}

func (xt XTunnel) LocalAddress() string {
	return xt.localListener.Addr().String()
}

// Pipe creates a full-duplex pipe between the two sockets and transfers data from one to the other.
func Pipe(conn1 net.Conn, conn2 net.Conn) {
	chan1 := chanFromConn(conn1)
	chan2 := chanFromConn(conn2)

	for {
		select {
		case b1 := <-chan1:
			if b1 != nil {
				conn2.Write(b1)
			} else {
				return
			}
		case b2 := <-chan2:
			if b2 != nil {
				conn1.Write(b2)
			} else {
				return
			}
		}
	}
}

// chanFromConn creates a channel from a Conn object, and sends everything it
//  Read()s from the socket to the channel.
func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)

	go func() {
		b := make([]byte, 1024)

		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				// Copy the buffer so it doesn't get changed while read by the recipient.
				copy(res, b[:n])
				c <- res
			}
			if err != nil {
				c <- nil
				break
			}
		}
	}()

	return c
}
