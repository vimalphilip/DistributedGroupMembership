package main

import (

)

/*
Called by introducer after receiving a join message. Compares timestamp of the message and.
the membershiplist. If timestamp in memberhiplist is more recent than message time, return 1
 otherwise return 0
*/

func checkTimeStamp(m member) int {
	return 0
}


//Helper function to write membershipList to file
func writeMLtoFile(){
	
}

//Function for introducer to send "isAlive" messages to VM's in it's membershiplist after reboot
//This is to check validity of local membershipList in introducer, as and when introducer crashes and needs to restart
func checkMLValid(){

}