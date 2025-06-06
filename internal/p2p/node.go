package p2p

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

type Node struct {
	Host      peer.ID
	topic     *pubsub.Topic
	sub       *pubsub.Subscription
	multiaddr string
}

const Rendezvous = "decentralized-rate-limiter"

func NewNode(ctx context.Context, topicName string) *Node {
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
	setupPeerDiscovery(ctx, h)
	return &Node{
		Host:      h.ID(),
		topic:     topic,
		sub:       sub,
		multiaddr: address,
	}
}

func setupPeerDiscovery(ctx context.Context, h host.Host) {
	kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		panic(err)
	}
	connectToBootstrapPeer(ctx, h)
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
}

func connectToBootstrapPeer(ctx context.Context, h host.Host) {
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

func (n *Node) Broadcast(data []byte) error {
	return n.topic.Publish(context.Background(), data)
}

func (n *Node) ReadLoop(handle func([]byte)) {
	go func() {
		for {
			msg, err := n.sub.Next(context.Background())
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
