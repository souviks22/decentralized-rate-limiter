package internal

import (
	"context"
	"log"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	multiaddr "github.com/multiformats/go-multiaddr"
)

type P2PNode struct {
	Host      peer.ID              `json:"host"`
	Topic     *pubsub.Topic        `json:"topic"`
	Sub       *pubsub.Subscription `json:"sub"`
	Multiaddr string               `json:"multiaddr"`
	DHT       *dht.IpfsDHT         `json:"dht"`
}

const Rendezvous = "decentralized-rate-limiter"

func NewP2PNode(ctx context.Context, topicName string) *P2PNode {
	host, err := libp2p.New()
	if err != nil {
		panic(err)
	}
	gossip, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(err)
	}
	topic, err := gossip.Join(topicName)
	if err != nil {
		panic(err)
	}
	sub, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}
	multiaddr := host.Addrs()[4].String() + "/p2p/" + host.ID().String()
	dht := SetupPeerDiscovery(ctx, host)
	log.Println("P2P node started at:", multiaddr)
	return &P2PNode{
		Host:      host.ID(),
		Topic:     topic,
		Sub:       sub,
		Multiaddr: multiaddr,
		DHT:       dht,
	}
}

func SetupPeerDiscovery(ctx context.Context, host host.Host) *dht.IpfsDHT {
	kadDht, err := dht.New(ctx, host, dht.Mode(dht.ModeServer))
	if err != nil {
		panic(err)
	}
	ConnectToBootstrapPeer(ctx, host)
	err = kadDht.Bootstrap(ctx)
	if err != nil {
		panic(err)
	}
	go ConnectToNearbyPeers(ctx, host, kadDht)
	return kadDht
}

func ConnectToBootstrapPeer(ctx context.Context, host host.Host) {
	bootstrap := os.Getenv("BOOTSTRAP_PEER")
	if bootstrap == "" {
		return
	}
	multiaddr, err := multiaddr.NewMultiaddr(bootstrap)
	if err != nil {
		panic(err)
	}
	bootPeer, err := peer.AddrInfoFromP2pAddr(multiaddr)
	if err != nil {
		panic(err)
	}
	err = host.Connect(ctx, *bootPeer)
	if err != nil {
		panic(err)
	}
	log.Println("Connected to:", bootPeer.ID)
}

func ConnectToNearbyPeers(ctx context.Context, host host.Host, kadDht *dht.IpfsDHT) {
	routingDiscovery := routing.NewRoutingDiscovery(kadDht)
	routingDiscovery.Advertise(ctx, Rendezvous)
	peers, _ := routingDiscovery.FindPeers(ctx, Rendezvous)
	for peer := range peers {
		if peer.ID == host.ID() {
			continue
		}
		host.Connect(ctx, peer)
	}
}

func (node *P2PNode) Broadcast(data []byte) error {
	return node.Topic.Publish(context.Background(), data)
}

func (node *P2PNode) ReadLoop(callback func([]byte)) {
	for {
		message, err := node.Sub.Next(context.Background())
		if err != nil {
			continue
		}
		if message.ReceivedFrom == node.Host {
			continue
		}
		callback(message.Data)
	}
}
