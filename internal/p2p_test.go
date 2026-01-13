package internal

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestConnectivity(t *testing.T) {
	bootstrap, node := createNetwork()
	if bootstrap.DHT.RoutingTable().Size() != 1 || node.DHT.RoutingTable().Size() != 1 {
		t.Error("Node not connected")
	}
}

func TestGossipMessaging(t *testing.T) {
	bootstrap, node := createNetwork()
	go bootstrap.ReadLoop(func(b []byte) {
		for i := range b {
			if b[i] != byte(i) {
				t.Error("Data corrupted")
			}
		}
	})
	node.Broadcast([]byte{0, 1, 2, 3, 4})
}

func createNetwork() (*P2PNode, *P2PNode) {
	bootstrap := NewP2PNode(context.Background(), "test")
	time.Sleep(5 * time.Second)
	os.Setenv("BOOTSTRAP_PEER", bootstrap.Multiaddr)
	node := NewP2PNode(context.Background(), "test")
	os.Unsetenv("BOOTSTRAP_PEER")
	time.Sleep(5 * time.Second)
	return bootstrap, node
}
