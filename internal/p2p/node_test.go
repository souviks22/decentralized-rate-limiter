package p2p

import (
	"context"
	"os"
	"testing"
	"time"
)

func createNetwork() (*Node, *Node) {
	bootstrap := NewNode(context.Background(), "test")
	time.Sleep(5 * time.Second)
	os.Setenv("BOOTSTRAP_PEER", bootstrap.multiaddr)
	node := NewNode(context.Background(), "test")
	os.Unsetenv("BOOTSTRAP_PEER")
	time.Sleep(5 * time.Second)
	return bootstrap, node
}

func TestConnectivity(t *testing.T) {
	bootstrap, node := createNetwork()
	if bootstrap.dht.RoutingTable().Size() != 1 || node.dht.RoutingTable().Size() != 1 {
		t.Error("Node not connected")
	}
}

func TestGossipMessaging(t *testing.T) {
	bootstrap, node := createNetwork()
	bootstrap.ReadLoop(func(b []byte) {
		for i := range b {
			if b[i] != byte(i) {
				t.Error("Data corrupted")
			}
		}
	})
	node.Broadcast([]byte{0, 1, 2, 3, 4})
}
