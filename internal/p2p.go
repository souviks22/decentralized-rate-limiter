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
	Host      peer.ID
	Topic     *pubsub.Topic
	Sub       *pubsub.Subscription
	Multiaddr string
	DHT       *dht.IpfsDHT
}

const Rendezvous = "decentralized-rate-limiter"

func NewP2PNode(ctx context.Context, topicName string) *P2PNode {
	h, err := libp2p.New()
	if err != nil {
		panic(err)
	}
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}
	topic, err := ps.Join(topicName)
	if err != nil {
		panic(err)
	}
	sub, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}
	address := h.Addrs()[4].String() + "/p2p/" + h.ID().String()
	log.Println("P2P node started at:", address)
	kadDHT := SetupPeerDiscovery(ctx, h)
	return &P2PNode{
		Host:      h.ID(),
		Topic:     topic,
		Sub:       sub,
		Multiaddr: address,
		DHT:       kadDHT,
	}
}

func SetupPeerDiscovery(ctx context.Context, h host.Host) *dht.IpfsDHT {
	kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		panic(err)
	}
	ConnectToBootstrapPeer(ctx, h)
	err = kadDHT.Bootstrap(ctx)
	if err != nil {
		panic(err)
	}
	routingDiscovery := routing.NewRoutingDiscovery(kadDHT)
	routingDiscovery.Advertise(ctx, Rendezvous)
	go func() {
		peerChan, _ := routingDiscovery.FindPeers(ctx, Rendezvous)
		for p := range peerChan {
			if p.ID == h.ID() {
				continue
			}
			h.Connect(ctx, p)
		}
	}()
	return kadDHT
}

func ConnectToBootstrapPeer(ctx context.Context, h host.Host) {
	bootstrapPeer := os.Getenv("BOOTSTRAP_PEER")
	if bootstrapPeer == "" {
		return
	}
	addr, err := multiaddr.NewMultiaddr(bootstrapPeer)
	if err != nil {
		panic(err)
	}
	p, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		panic(err)
	}
	err = h.Connect(ctx, *p)
	if err != nil {
		panic(err)
	}
	log.Println("Connected to:", p.ID)
}

func (n *P2PNode) Broadcast(data []byte) error {
	return n.Topic.Publish(context.Background(), data)
}

func (n *P2PNode) ReadLoop(handle func([]byte)) {
	go func() {
		for {
			msg, err := n.Sub.Next(context.Background())
			if err != nil {
				continue
			}
			if msg.ReceivedFrom == n.Host {
				continue
			}
			handle(msg.Data)
		}
	}()
}
