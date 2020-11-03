package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/jjg-akers/Control-Plane/cmd/link"
	"github.com/jjg-akers/Control-Plane/cmd/network"
)

//Settings
var (
	hostQueueSize   = 1000
	routerQueueSize = 1000 // 0 means unlimited
	simulationTime  = 2    // give the network sufficient time to transfer all packets before quitting
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
	costDRouterA := make(map[string]map[int]int)
	costDRouterA["H1"] = map[int]int{0:1}
	costDRouterA["RB"] = map[int]int{1:1}
	routerA := network.NewRouter("RA", costDRouterA, routerQueueSize)
	objectL = append(objectL, routerA)


	costDRouterB := make(map[string]map[int]int)
	costDRouterB["H2"] = map[int]int{1:3}
	costDRouterB["RA"] = map[int]int{0:1}
	routerB := network.NewRouter("RB", costDRouterB, routerQueueSize)
	objectL = append(objectL, routerB)

	// create a link layer to keep track of links between network nodes
	linkLayer := link.NewLinkLayer()
	objectL = append(objectL, linkLayer)

	// add all the links -- need to reflect the connectivity in costD tables above
	linkLayer.AddLink(link.NewLink(host1, 0, routerA, 0))
	linkLayer.AddLink(link.NewLink(routerA, 1, routerB, 0))
	linkLayer.AddLink(link.NewLink(routerB, 1, host2, 0))


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
	routerA.SendRoutes(1)	// one update starts the routing procss
	time.Sleep(time.Duration(simulationTime) * time.Second)

	fmt.Println("converged routing tables")
	for _, obj := range objectL{
		switch v := obj.(type) {
		case *network.Router:
			v.PrintRoutes()
		default:
		}
	}
	
	// send packet form host 1 to host 2
	host1.UdtSend("H2", "MESSAGE_FROM_H!")
	time.Sleep(time.Duration(simulationTime) * time.Second)



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