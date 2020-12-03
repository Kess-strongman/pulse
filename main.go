package main

import (
	"fmt"
	"httpingest/config"
	"httpingest/utils"
	"log"
)

func main() {

	configLoadErr := config.Config.LoadFromFile("assets/config.json")
	if configLoadErr != nil {
		fmt.Println("Could not get configuration")
		return
	}
	utils.Logging()
	fmt.Println(config.Config.GetConnString())
	config.InitDatabase(config.Config.GetConnString())
	service := NewServer(config.Config.GetPort())

	// Run any workers here
	//service.WG.Add(1)

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

	//wait shutdown
	service.WaitShutdown()

	log.Printf("Service Exiting")
}
