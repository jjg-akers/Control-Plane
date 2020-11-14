package main

import (
	"fmt"
	"sync"
	"time"

	link "github.com/jjg-akers/Control-Plane/cmd/link/link3"
	network "github.com/jjg-akers/Control-Plane/cmd/network/network3"
)

//Settings
var (
	hostQueueSize   = 1000
	routerQueueSize = 1000 // 0 means unlimited
	simulationTime  = 15   // give the network sufficient time to transfer all packets before quitting
	wg              = sync.WaitGroup{}
)

func main() {

	//keep track of objects so we can kill their threads
	objectL := []interface{}{}
	// hostL := []*network.Host{}
	// routerL := []*network.Router{}

	// create network nodes
	host1 := network.NewHost("H1", hostQueueSize)
	objectL = append(objectL, host1)

	host2 := network.NewHost("H2", hostQueueSize)
	objectL = append(objectL, host2)

	//Create Routers
	// costD: cost table to neighbors {neighbor: {interface: cost}}

	costDRouterA := make(map[string][2]int)
	costDRouterA["H1"] = [2]int{0, 1}
	costDRouterA["RB"] = [2]int{1, 1}
	costDRouterA["RC"] = [2]int{2, 5}
	routerA := network.NewRouter("RA", costDRouterA, routerQueueSize)
	objectL = append(objectL, routerA)

	costDRouterB := make(map[string][2]int)
	costDRouterB["RA"] = [2]int{0, 1}
	costDRouterB["RD"] = [2]int{1, 1}
	routerB := network.NewRouter("RB", costDRouterB, routerQueueSize)
	objectL = append(objectL, routerB)

	costDRouterC := make(map[string][2]int)
	costDRouterC["RA"] = [2]int{0, 5}
	costDRouterC["RD"] = [2]int{1, 1}
	routerC := network.NewRouter("RC", costDRouterC, routerQueueSize)
	objectL = append(objectL, routerC)

	costDRouterD := make(map[string][2]int)
	costDRouterD["RB"] = [2]int{0, 10}
	costDRouterD["RC"] = [2]int{1, 1}
	costDRouterD["H2"] = [2]int{2, 1}
	routerD := network.NewRouter("RD", costDRouterD, routerQueueSize)
	objectL = append(objectL, routerD)

	// create a link layer to keep track of links between network nodes
	linkLayer := link.NewLinkLayer()
	objectL = append(objectL, linkLayer)

	// add all the links -- need to reflect the connectivity in costD tables above
	linkLayer.AddLink(link.NewLink(host1, 0, routerA, 0))

	linkLayer.AddLink(link.NewLink(routerA, 1, routerB, 0))
	linkLayer.AddLink(link.NewLink(routerA, 2, routerC, 0))

	linkLayer.AddLink(link.NewLink(routerB, 1, routerD, 0))

	linkLayer.AddLink(link.NewLink(routerC, 1, routerD, 1))

	linkLayer.AddLink(link.NewLink(routerD, 2, host2, 0))

	// start all the objects
	for _, obj := range objectL {
		switch v := obj.(type) {
		case *network.Host:
			v.Run(&wg)
		case *network.Router:
			v.Run(&wg)
		case *link.LinkLayer:
			v.Run(&wg)

		default:
			fmt.Printf("type: %T, value: %v\n", v, v)
			fmt.Println("default")
		}
	}

	// compute routing tables
	routerA.SendRoutes(1) // one update starts the routing procss
	time.Sleep(time.Duration(simulationTime) * time.Second)

	fmt.Println("CONVERGED ROUTING TABLES:")
	for _, obj := range objectL {
		switch v := obj.(type) {
		case *network.Router:
			v.PrintRoutes()
		default:
		}
	}

	fmt.Println("Start sending messages")
	fmt.Println("")

	// send packet form host 1 to host 2
	host1.UdtSend("H2", "MESSAGE_FROM_H1")
	time.Sleep(2 * time.Second)
	host2.UdtSend("H1", "MESSAGE_FROM_H2")
	time.Sleep(2 * time.Second)

	fmt.Println("")

	fmt.Println("done sending messages")
	fmt.Println("")
	// create some events
	// i := 0
	// for i < 3 {
	// 	client.UdtSend(2, fmt.Sprintf("Sample data %d", i))
	// 	i++
	// }

	// give the network sufficient time to transfer all packets before quitting
	//time.Sleep(time.Duration(simulationTime) * time.Second)

	// join all thread
	for _, obj := range objectL {
		switch v := obj.(type) {
		case *network.Host:
			v.Stop <- true
		case *network.Router:
			v.Stop <- true
		case *link.LinkLayer:
			v.Stop <- true

		default:
			fmt.Printf("type: %T, value: %v\n", v, v)
			fmt.Println("default")
		}
	}

	// send the stop signal and wait
	fmt.Println("waiting")
	wg.Wait()

	fmt.Println("done waiting")
	// need to wait here for routines
}
