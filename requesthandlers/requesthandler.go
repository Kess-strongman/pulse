package requesthandlers

import (
	"fmt"
	"net/http"
	"time"
)

func lala() {
	var malarkey time.Time
	fmt.Println("Woop", malarkey)

}

// ServiceGetStatusHandler - receives a request for service status
func ServiceGetStatusHandler(w http.ResponseWriter, r *http.Request) {
	// First we check that the request is valid - Don't worry at this stage we will implement this next week

	// Then once it's authorised - We get the service status from the Database

}
