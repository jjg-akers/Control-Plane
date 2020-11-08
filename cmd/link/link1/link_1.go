package link

import (
	"fmt"
	"log"
	"sync"

	network "github.com/jjg-akers/Control-Plane/cmd/network/network1"
)

//Creates a link between two objects by looking up and linking node interfaces
// from_node: network Host from which data will be transfered
// from_intf_num: number of the interface on that node
// to_node: network Host to which data will be transfered
// to_intf_num: number of the interface on that node
// mtu: link maximum transmission unit

//Link is An abstraction of a link between router interfaces
type Link struct {
	Node1          network.Node
	Node1Interface int
	Node2          network.Node
	Node2Interface int
	// InIntf      *network.NetworkInterface
	// OutIntf     *network.NetworkInterface
}

//NewLink return a new link
func NewLink(node1 network.Node, node1Intf int, node2 network.Node, node2Intf int) *Link {
	toReturn := Link{
		Node1:          node1,
		Node1Interface: node1Intf,
		Node2:          node2,
		Node2Interface: node2Intf,
	}

	return &toReturn
}

func (lk *Link) Str() string {
	return fmt.Sprintf("Link %s-%d to %s-%d", lk.Node1.Str(), lk.Node1Interface, lk.Node2.Str(), lk.Node2Interface)
}

//txPkt transmits a packet between interfaces in each direction
func (lk *Link) TxPkt() {

	for i := 0; i < 2; i++ {
		var (
			nodeA     network.Node
			nodeAIntf int
			nodeB     network.Node
			nodeBIntf int
		)

		if i == 0 {
			nodeA = lk.Node1
			nodeAIntf = lk.Node1Interface
			nodeB = lk.Node2
			nodeBIntf = lk.Node2Interface
		} else {
			nodeA = lk.Node2
			nodeAIntf = lk.Node2Interface
			nodeB = lk.Node1
			nodeBIntf = lk.Node1Interface
		}

		intfA := nodeA.GetInterfaceL()[nodeAIntf]
		intfB := nodeB.GetInterfaceL()[nodeBIntf]

		pktS, err := intfA.Get("out")
		if err != nil {
			continue
		}

		err = intfB.Put(pktS, "in", false)
		if err != nil {
			fmt.Printf("%s: direction %s-%d -> %s-%d: packet lost\n", lk.Str(), nodeA.Str(), nodeAIntf, nodeB.Str(), nodeBIntf)
		} else {
			fmt.Printf("%s: direction %s-%d -> %s-%d: transmitting packet %s\n", lk.Str(), nodeA.Str(), nodeAIntf, nodeB.Str(), nodeBIntf, pktS)
		}
	}
}

//LinkLayer is a list of links in the network
// Stop is for routine termination
type LinkLayer struct {
	LinkL []*Link
	Stop  chan interface{}
}

func NewLinkLayer() *LinkLayer {
	return &LinkLayer{
		LinkL: []*Link{},
		Stop:  make(chan interface{}, 1),
	}
}

//Str returns the name of the network layer
func (ll *LinkLayer) Str() string {
	return "Network"
}

//AddLink add a link to the network
func (ll *LinkLayer) AddLink(lk *Link) {
	ll.LinkL = append(ll.LinkL, lk)
}

//Transfer transfers a packet across all links
func (ll *LinkLayer) Transfer() {
	for _, link := range ll.LinkL {
		link.TxPkt()
	}
}

//Run starts a routing for the network to keep transmitting data across links
func (ll *LinkLayer) Run(wg *sync.WaitGroup) {
	log.Println("LinkLayer 'Run' routine starting")
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-ll.Stop:
				log.Println("linklayer got close signal")
				log.Println("LinkLayer 'Run' routine ending")
				return
			default:
				// transfer one packet on all the links
				ll.Transfer()
			}
		}
	}(wg)
}
