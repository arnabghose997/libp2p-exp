package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	ipfsutil "berty.tech/weshnet/pkg/ipfsutil"
	ipfs_mobile "berty.tech/weshnet/pkg/ipfsutil/mobile"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	ifacecore "github.com/ipfs/interface-go-ipfs-core"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	libp2p_host "github.com/libp2p/go-libp2p/core/host"
)

const swarmKeyFile = "swarm.key"

// someHostingFunc is a placeholder for libp2p host configuration
func someHostingFunc(id peer.ID, ps peerstore.Peerstore, options ...libp2p.Option) (libp2p_host.Host, error) {
	return libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000",         // regular tcp connections
			"/ip4/0.0.0.0/udp/9000/quic-v1",        // regular tcp connections
		),
		libp2p.DefaultTransports,
		libp2p.NATPortMap(),
	)
}

// createRepoPath checks if the repo path exists, and creates it if it doesn't
func createRepoPath(repoPath string) error {
	

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		fmt.Println("Repo path does not exist, creating...")
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			return err
		}
	}
	return nil
}

// generateSwarmKey generates or loads a swarm key
func generateSwarmKey(repoPath string) (string, error) {
	// Ensure the repo path exists
	if err := createRepoPath(repoPath); err != nil {
		return "", err
	}

	keyPath := filepath.Join(repoPath, swarmKeyFile)
	if _, err := os.Stat(keyPath); err == nil {
		fmt.Println("Swarm key exists, loading...")
		data, err := os.ReadFile(keyPath)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	fmt.Println("Swarm key not found, generating new one...")
	newKey := "/key/swarm/psk/1.0.0/\n/base16/\n278b9a199c43fa84178920bd9f5cbcd69e933ddf02a8f69e47a3ea5a1705512f"
	if err := os.WriteFile(keyPath, []byte(newKey), 0600); err != nil {
		return "", err
	}

	return newKey, nil
}

// createNode creates and returns an IPFS CoreAPI instance
func createNode(repoPath string) (ipfsutil.ExtendedCoreAPI, error) {
	ipfsrepo, err := ipfsutil.LoadRepoFromPath(repoPath)
	if err != nil {
		return nil, err
	}

	mrepo := ipfs_mobile.NewRepoMobile(repoPath, ipfsrepo)
	mnode, err := ipfsutil.NewIPFSMobile(context.TODO(), mrepo, &ipfsutil.MobileOptions{
		HostOption: someHostingFunc,
	})
	if err != nil {
		return nil, err
	}

	return ipfsutil.NewExtendedCoreAPIFromNode(mnode.IpfsNode)
}

// checkPeerAvailability simulates the `ipfs ping` command
func checkPeerAvailability(api ifacecore.CoreAPI, peerID string) (string, error) {
	pid, err := peer.Decode(peerID)
	if err != nil {
		return "", fmt.Errorf("failed to decode peer ID: %w", err)
	}

	peerInfo, err := api.Dht().FindPeer(context.Background(), pid)
	if err != nil {
		return "Ping failed: peer unreachable", nil
	}

	return fmt.Sprintf("Ping successful! Peer info: %v", peerInfo.ID), nil
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("IPFS Mobile Node")

	// Swarm key label, initially hidden
	swarmKeyLabel := widget.NewLabel("Swarm Key: ")
	swarmKeyLabel.Hide()

	peerIDEntry := widget.NewEntry()
	peerIDEntry.SetPlaceHolder("Enter PeerID of counterparty IPFS node")

	statusLabel := widget.NewLabel("Status: Waiting for action")

	// Check Peer button, initially disabled
	pingButton := widget.NewButton("Check Peer", nil)
	pingButton.Disable()

	var ipfsAPI ifacecore.CoreAPI

	// Start button to initialize the IPFS node
	baseStoragePath := fyne.CurrentApp().Storage().RootURI().Path()
	repoPath := filepath.Join(baseStoragePath, ".ipfs")

	startButton := widget.NewButton("Start", func() {
		// Generate or load the swarm key
		swarmKey, err := generateSwarmKey(repoPath)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Failed to generate/load swarm key: %v", err))
			return
		}

		// Create IPFS node
		ipfsAPI, err = createNode(repoPath)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Failed to create IPFS node: %v", err))
			return
		}

		// Display swarm key and enable the Check Peer button
		swarmKeyLabel.SetText("Swarm Key: " + swarmKey)
		swarmKeyLabel.Show()
		pingButton.Enable()
		statusLabel.SetText("IPFS node started successfully")
	})

	// Functionality for the Check Peer button
	pingButton.OnTapped = func() {
		peerID := peerIDEntry.Text
		if peerID == "" {
			statusLabel.SetText("Please enter a PeerID")
			return
		}

		// Check peer availability
		result, err := checkPeerAvailability(ipfsAPI, peerID)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Check failed: %v", err))
			return
		}
		statusLabel.SetText(result)
	}

	content := container.NewVBox(
		swarmKeyLabel,
		peerIDEntry,
		startButton,
		pingButton,
		statusLabel,
	)
	scroll := container.NewScroll(content)

	myWindow.SetContent(scroll)

	myWindow.Resize(fyne.NewSize(300, 400))
	myWindow.ShowAndRun()
}
