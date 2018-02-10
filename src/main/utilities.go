package main

import (
	
)

//utility or helper methods

//Initialize membershipList with current time and LocalIP
func initializeML(){

}


func updateML(){

}

//get index for local VM in membershipList
func getIndex() int {
	for i, element := range membershipList {
		if currHost == element.host {
			return i
		}
	}
	return -1
}


//Logging
//Helper function to log errors
func errorCheck(err error) {
	if err != nil {
		errlog.Println(err)
	}
}

//Helper function to log general information
func infoCheck(info string) {
	infolog.Println(info)
}

//Helper function to hard reset both timers (stop both and set resetFlags to 1)
func resetTimers() {
	resetFlags[0] = 1
	resetFlags[1] = 1
	timers[0].Reset(0)
	timers[1].Reset(0)
}