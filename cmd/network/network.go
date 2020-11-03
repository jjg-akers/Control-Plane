package network

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
)

type Node interface {
	Run(*sync.WaitGroup)
	GetInterfaceL() []*NetworkInterface
	Str() string
}

// Global setting variables
var dstAddrStrLength = 5
var protSLength = 1

//Wrapper class for a queue of packets
// param maxsize - the max size of the queue storing packets
type NetworkInterface struct {
	muIn  sync.Mutex
	muOut sync.Mutex
	QueueIn  []string
	QueueOut []string
	MaxQueSize int
}

func NewNetworkInterface(maxQ int) *NetworkInterface {
	return &NetworkInterface{
		muIn:      sync.Mutex{},
		muOut:      sync.Mutex{},
		QueueIn: []string{},
		QueueOut: []string{},
		MaxQueSize: maxQ,
	}
}

//gets a packet from the queue
// returns an error if the 'queue' is empty
func (n *NetworkInterface) Get(inORout string) (string, error) {

	if inORout == "in" {
		n.muIn.Lock()
		defer n.muIn.Unlock()
		if len(n.QueueIn) > 0 {
			toReturn := n.QueueIn[0]
			n.QueueIn = n.QueueIn[1:]
			return toReturn, nil
		}
		return "", errors.New("Empty")
	}

	if len(n.QueueOut) > 0 {
		n.muOut.Lock()
		defer n.muOut.Unlock()
		toReturn := n.QueueOut[0]
		n.QueueOut = n.QueueOut[1:]
		return toReturn, nil
	}
	return "", errors.New("Empty")
}

//put the packet into the queue
// put returns an error if the queue is full
func (n *NetworkInterface) Put(pkt, inORout string, block bool) error {
	// if block is true, block until there is room in the queue
	// if false, throw queue full error
	if block == true {

		if inORout == "in" {
			for {
				// obtain lock
				n.muIn.Lock()
				if len(n.QueueIn) < n.MaxQueSize {
					// add to queue
					n.QueueIn = append(n.QueueIn, pkt)
					n.muIn.Unlock()
					return nil
				}
				// unlock until next loop
				n.muIn.Unlock()
				continue
			}
		}

		for {
			// obtain lock
			n.muOut.Lock()
			if len(n.QueueOut) < n.MaxQueSize {
				// add to queueOut
				n.QueueOut = append(n.QueueOut, pkt)
				n.muOut.Unlock()
				return nil
			}
			// unlock until next loop
			n.muOut.Unlock()
			continue
		}

	}

	// if block != true
	if inORout == "in" {
		n.muIn.Lock()
		defer n.muIn.Unlock()
		if len(n.QueueIn) < n.MaxQueSize {
			n.QueueIn = append(n.QueueIn, pkt)
			return nil
		}

		return errors.New("QueueIn Full")
	}
	n.muOut.Lock()
	defer n.muOut.Unlock()
	if len(n.QueuOt) < n.MaxQueSize {
		n.QueuOt = append(n.QueuOt, pkt)
		return nil
	}

	return errors.New("QueuOt Full")
}

//Implements a network layer packet
// DstAddr: address of the destination host
// DataS: packet payload
// DstAddrStrLength: packet encoding lengths
type NetworkPacket struct {
	DstAddr          int
	DataS            string
	ProtS            string
	DstAddrStrLength int
}

func NewNetworkPacket(dstAddr int, protS, dataS string) *NetworkPacket {
	return &NetworkPacket{
		DstAddr:          dstAddr,
		DataS:            dataS,
		ProtS:            protS,
		DstAddrStrLength: dstAddrStrLength,
	}
}

func (np *NetworkPacket) Str() string {
	return np.ToByteS()
}

//ToBytesS converts packet to a byte string for transmission over links
func (np *NetworkPacket) ToByteS() (string, error) {
	byteS := fmt.Sprintf("%0*s", np.DstAddrStrLength, strconv.Itoa(np.DstAddr))
	if np.ProtS == "data" {
		byteS += "1"
	} else if np.ProtS == "control" {
		byteS += "2"
	} else {
		return "", fmt.Errorf("%s: unknown protS option: %s\n", np.Str(), np.ProtS))
	}

	byteS += np.DataS
	return byteS
	//seqNumS := fmt.Sprintf("%0*s", p.SeqNumSlength, strconv.Itoa(p.SeqNum))
}

//FromByteS builds a packet object from a byte string
// Returns error if it cannot convert addres to int
func FromByteS(byteS string) (*NetworkPacket, error) {
	dstAddr, err := strconv.Atoi(byteS[0:dstAddrStrLength])
	if err != nil {
		log.Println("Error converting addr to string")
		return nil, err
	}
	protS = byteS[dstAddrStrLength:dstAddrStrLength+protSLength]
	if protS == "1" {
		protS = "data"
	} else if protS == "2" {
		protS = "control"
	} else {
		return "", fmt.Errorf("Unknown protS option: %s\n", protS))
	}

	dataS := byteS[dstAddrStrLength + protSLength:]
	return NewNetworkPacket(dstAddr, protS, dataS), nil
}

//Host implements a network host for receiving and transmitting data
// Addr: address of this node represented as an integer
type Host struct {
	Addr          int
	InterfaceL  []*NetworkInterface
	// OutInterfaceL []*NetworkInterface
	Stop          chan interface{}
}

func (h *Host) GetInterfaceL() []*NetworkInterface {
	return h.InterfaceL
}

// func (h *Host) GetInInterfaceL() []*NetworkInterface {
// 	return h.InInterfaceL
// }

// func (h *Host) GetOutInterfaceL() []*NetworkInterface {
// 	return h.OutInterfaceL
// }

func NewHost(addr int, maxQSize int) *Host {
	return &Host{
		Addr:          addr,
		InterfaceL:  []*NetworkInterface{NewNetworkInterface(maxQSize)},
		Stop:          make(chan interface{}, 1),
	}
}

// Called when printing the objects
func (h *Host) Str() string {
	return fmt.Sprintf("Host_%d", h.Addr)
}

//UdtSend creates a packet and enqueues for transmission
// dst_addr: destination address for the packet
// data_S: data being transmitted to the network layer
func (h *Host) UdtSend(dstAddr int, dataS string) {
	p := NewNetworkPacket(dstAddr, dataS)

	fmt.Printf("%s: sending packet \"%s\"\n", h.Str(), p.ToByteS())

	err := h.InterfaceL[0].Put(p.ToByteS(), "out", false) // send packets always enqueued successfully
	if err != nil {
		fmt.Println("err from put in UDTsent: ", err)
	}
}

//UdtReceive receives packest from the network layer
func (h *Host) UdtReceive() {
	pktS, err := h.InterfaceL[0].Get("in")
	if err == nil {
		fmt.Printf("%s: received packet \"%s\"\n", h.Str(), pktS)
	}
}

//Run startes a routine for the host to keep receiving data
func (h *Host) Run(wg *sync.WaitGroup) {
	fmt.Println("Starting host receive routine")
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-h.Stop:
				log.Println("Host got close signal")
				fmt.Println("Ending host receive routine")
				return
			default:
				//receive data arriving in the in interface
				h.UdtReceive()
			}
		}
	}(wg)
}

//Router implements a multi-interface router described in class
// Name: friendly router nam for debugging
// costD: cost table to neighbors {neighbor: {interface: cost}}
type Router struct {
	Stop          chan interface{}
	Name          string
	InterfaceL  []*NetworkInterface
	CostD	map[string]map[int]int
	RtTableD map[string]map[int]int
}

//NewRouter returns a new router with given specs
// interfaceCount: the number of input and output interfaces
// maxQueSize: max queue legth (passed to interfacess)

// rounter needs to implement packet segmentation is packet is too big for interface
func NewRouter(name string, costD, interfaceCount int, maxQueSize int) *Router {
	in := make([]*NetworkInterface, interfaceCount)
	for i := 0; i < interfaceCount; i++ {
		in[i] = NewNetworkInterface(maxQueSize)
	}

	return &Router{
		Stop:          make(chan interface{}, 1),
		Name:          name,
		InterfaceL:  in,
		CostD: costD,
	}
}

func (rt *Router) GetInterfaceL() []*NetworkInterface {
	return rt.InterfaceL
}

// func (rt *Router) GetOutInterfaceL() []*NetworkInterface {
// 	return rt.OutInterfaceL
// }

//Called when printing the object
func (rt *Router) Str() string {
	return fmt.Sprintf("Router_%s", rt.Name)
}

func (rt *Router) PrintRoutes(){
	// TODO print the routes as a two dimensional table
	fmt.Println(rt.RtTableD)
}

//lok through the content of incoming interfaces and forward to appropriate outgoing interfaces
func (rt *Router) processQueues() {
	for i, v := range rt.InterfaceL {
		//pktS := ""

		// TRYE
		// get packet from interface i
		if pktS, err := v.Get("in"); err == nil {
			//fmt.Println("in routher forward, packet from Get(): ", pktS)
			// if packet exists make a forwarding decision
			p, err := FromByteS(pktS)
			if err != nil {
				log.Println("Could not get packet")
				continue
			}

			if p.ProtS == "data"{
				rt.forwardPacket(p, i)
			} else if p.protS == "control"{
				rt.updateRoutes(p,i)
			} else{
				// Raise Error
				log.Printf("%s: unknown packet type in packet %s\n", rt.Str(), p.Str())
			}

		}
		//log.Println("no packet to forard in router")
	}
}


func (rt *Router) forwardPacket(p *NetworkPacket, i int){
			// HERE you will need to implement a lookup into the
			// forwarding table to find the appropriate outgoing interface
			// for now we assume the outgoing interface is 1
			if err = rt.InterfaceL[1].Put(p.ToByteS(), "out", true); err != nil {
				//log.Printf("Could not put packet %s in router %s, into outInterface %d. Error: %s", p.str, rt.forward, i, err)
				log.Printf("%s: packet '%s' lost on interface %d\n", rt.Str(), p.Str(), i)
			}
			
			fmt.Printf("%s: forwarding packet %s from interface %d to %d\n", rt.Str(), p.Str(), i, 1)
}

//SendRoutes will send out route updates
//	param i: Interface number on which to send out routing update
func (rt *Router) SendRoutes(i int){
	// TODO: Send out a routing table update
	
	// create a routing table update packet
	p := NewNetworkPacket(0, "control", "DUMMY_ROUTING_TABLE")

	fmt.Printf("%s: sending routing update \"%s\" from interface %d\n", rt.Str(), p.Str(), i)

	if err := rt.InterfaceL[i].Put(p.ToByteS(), "out", true); err != nil{
		fmt.Printf("%s: packet \"%s\" lost on interface %d\n", rt.Str(), p.Str(), i)
	}
}

//UpdateRoutes forwards the packet according to the routing table
// param p: packet containing routing information
func (rt *Router) UpdateRoutes(p *NetworkPacket, i int){
	// TODO: add logic to update the routing tables and possibly send out routing updates
	fmt.Printf("%s: Received routing update %s from interface %d\n", rt.Str(), p.Str(), i)

}

func (rt *Router) Run(wg *sync.WaitGroup) {
	fmt.Printf("%s: starting\n", rt.Str())

	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-rt.Stop:
				log.Println("router got close signal")
				fmt.Printf("%s: Ending\n", rt.Str())
				return
			default:
				rt.processQueues()
			}
		}
	}(wg)
}
