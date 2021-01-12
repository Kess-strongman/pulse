package main

import (
	"fmt"
	"log"
	"pulseservice/config"
	"pulseservice/utils"
	"time"
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
	//b, e := config.GetLatestEntries()
	//b, e := config.GetLatestMessageForApp("pulseTest")
	//b, e := config.GetLatestServiceStatusMessages()
	//b, e := config.GetLatestHelloMessages()
	sdate := "2021-01-07 13:53:24.86751"
	StartTime, StartTimeParseError := time.Parse("2006-01-02T15:04:05Z", sdate)
	if StartTimeParseError != nil {
		StartTime, StartTimeParseError = time.Parse("2006-01-02 15:04:05", sdate)

	}
	edate := "2021-01-07 13:56:25.022257"
	EndTime, EndTimeParseError := time.Parse("2006-01-02T15:04:05Z", edate)
	if EndTimeParseError != nil {
		EndTime, EndTimeParseError = time.Parse("2006-01-02 15:04:05", edate)

	}
	b, e := config.GetMessageForAppBetweenTimes("pulseTest", StartTime, EndTime)
	fmt.Println(b, e)

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
