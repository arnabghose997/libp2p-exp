package main

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	libp2p "github.com/libp2p/go-libp2p"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	ma "github.com/multiformats/go-multiaddr"
    "github.com/libp2p/go-libp2p/core/peer"
)


var node libp2pHost.Host

// Function to create a libp2p source node and return the host
func createSourceNode() libp2pHost.Host {
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/8008", // Use any port you like
		),
	)
	if err != nil {
		fmt.Println("Error creating node:", err)
		return nil
	}

	return host
}

// Function to get the Node PeerID
func getNodePeerID() string {
	if node != nil {
		return node.ID().String()
	} else {
		return "Error: Node is not running"
	}
}

// Function to ping another node by its multiaddress
func pingPeer(peerAddr string) string {
	// Parse the peer multiaddress
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/192.168.0.217/tcp/9000/p2p/%s", peerAddr))
	if err != nil {
		return fmt.Sprintf("Invalid address: %s", err)
	}

	// Extract peer ID from the address
	info, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Sprintf("Failed to extract peer info: %s", err)
	}

	// Connect to the peer
	if err := node.Connect(context.Background(), *info); err != nil {
		return fmt.Sprintf("Failed to connect: %s", err)
	}

	// Send a ping
	pingService := ping.NewPingService(node)
	res := pingService.Ping(context.Background(), info.ID)

	// Wait for the result
	select {
	case r := <-res:
		if r.Error == nil {
			return fmt.Sprintf("Ping successful! Latency: %s", r.RTT)
		} else {
			return fmt.Sprintf("Ping failed: %s", r.Error)
		}
	case <-time.After(time.Second * 10):
		return "Ping timeout"
	}
}

func main() {
	// Create a new application
	myApp := app.New()

	// Create the libp2p host (node)
	node = createSourceNode()
	defer func() {
		if node != nil {
			node.Close()
		}
	}()

	// Create a new window
	myWindow := myApp.NewWindow("Simple App")

	// Set the desired window size (width x height)
	myWindow.Resize(fyne.NewSize(400, 250))

	// Create an entry widget for input
	messageEntry := widget.NewEntry()
	messageEntry.SetPlaceHolder("Enter PeerID and Address")

	// Fetch the Node PeerID and create a label to display it
	peerID := getNodePeerID()
	peerIDLabel := widget.NewLabel(fmt.Sprintf("Node PeerID: %s", peerID))

	// Create a label to show the ping result
	resultLabel := widget.NewLabel("")

	// Create a button
	connectButton := widget.NewButton("Connect", func() {
		// Capture the input value
		peerAddr := messageEntry.Text

		// Attempt to ping the entered PeerID
		result := pingPeer(peerAddr)
		resultLabel.SetText(result)
		fmt.Println(result)
	})

	// Create a container to hold the widgets
	content := container.NewVBox(
		messageEntry,
		connectButton,
		peerIDLabel,  // Add the label to display PeerID
		resultLabel,  // Add the label to display the result
	)

	// Set the content of the window
	myWindow.SetContent(content)

	// Show and run the application
	myWindow.ShowAndRun()
}
