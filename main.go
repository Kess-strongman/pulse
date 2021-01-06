package main

import (
	"fmt"
	"log"
	"pulseservice/config"
	"pulseservice/utils"
)

func main() {
	// This is the setup section
	configLoadErr := config.Config.LoadFromFile("assets/config.json")
	if configLoadErr != nil {
		fmt.Println("Could not get configuration")
		return
	}
	approvedappsLoadErr := config.AA.LoadFromFile("assets/approvedapps.json")
	if approvedappsLoadErr != nil {
		fmt.Println("Could not get approved apps")
		return
	}
	utils.Logging()
	fmt.Println(config.Config.GetConnString())
	config.InitDatabase(config.Config.GetConnString())
	service := NewServer(config.Config.GetPort())
	// Worker Section
	// Run any workers here
	//service.WG.Add(1)

	// Start main service
	fmt.Printf("Starting Server on port %v\n", service.Addr)
	log.Printf("Starting Server on port %v\n", service.Addr)

	go func() {
		// This starts the HTTP server
		err := service.ListenAndServe()

		if err != nil {
			fmt.Println("Error: ", err)
			log.Fatalln("Cannot Start Server, exiting:", err.Error())
		}
	}()

	//Wait for shutdown
	service.WaitShutdown()
	// Do any final tidying up
	log.Printf("Service Exiting")
}
