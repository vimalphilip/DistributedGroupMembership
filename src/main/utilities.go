package main

import (
	"os"
	"time"
	"net"
	"log"
	
)

//utility or helper methods

//Initialize membershipList with current time and LocalIP
func initializeML(){
	node := member{currHost, time.Now().Format(time.RFC850)}
	membershipList = append(membershipList, node)
}


func updateML(){

}

//get index for local VM in membershipList
func getIndex() int {
	for i, element := range membershipList {
		if currHost == element.Host {
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

//Sets currHost to local IP (as a string)
//Sets membershipList with currHost as its only member with current time
//Initializes timers with MAX_TIME and subsequently stops them. This is to prevent false firing of timers when Syn/Ack begins
func setupAndInitialize() {
	currHost = getIP()
	initializeML()
	timers[0] = time.NewTimer(MAX_TIME)
	timers[1] = time.NewTimer(MAX_TIME)
	timers[0].Stop()
	timers[1].Stop()
	
	logfile, _ := os.OpenFile("logfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	errlog = log.New(logfile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	infolog = log.New(logfile, "INFO: ", log.Ldate|log.Ltime)

	
}

//get local IP address in the form of a string
func getIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		errorCheck(err)
	}
	return addrs[1].String()
}