package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/libp2p/go-libp2p/core/network"
)

// Function to create and run a libp2p node
func createNode() libp2pHost.Host {
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000", // Listen on port 9000
		),
	)
	if err != nil {
		log.Fatalf("Failed to create node: %s", err)
	}

	fmt.Println("Node running. PeerID:", host.ID())
	fmt.Println("Multiaddresses:")
	for _, addr := range host.Addrs() {
		fmt.Println(addr.Encapsulate(multiaddr.StringCast(fmt.Sprintf("/p2p/%s", host.ID()))))
	}

	// Set up a ping handler to log incoming pings
	pingService := ping.NewPingService(host)
	host.SetStreamHandler(ping.ID, func(stream network.Stream) {
		fmt.Println("Received a ping from:", stream.Conn().RemotePeer())

		// Handle the ping
		res := pingService.Ping(context.Background(), stream.Conn().RemotePeer())

		select {
		case r := <-res:
			if r.Error == nil {
				fmt.Printf("Ping latency: %s\n", r.RTT)
			} else {
				fmt.Println("Ping failed:", r.Error)
			}
		case <-time.After(time.Second * 10):
			fmt.Println("Ping timeout")
		}
		stream.Close()
	})

	return host
}

func main() {
	node := createNode()
	defer func() {
		if node != nil {
			node.Close()
		}
	}()

	// Wait for an interrupt signal to gracefully shut down the node
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	fmt.Println("Shutting down node...")
}
