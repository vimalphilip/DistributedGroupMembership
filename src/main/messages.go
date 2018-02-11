package main

import (
	
	"net"
	"fmt"
	"encoding/gob"
	"bytes"
	"time"
	

)

/*
 Handles connection protocol and writes message to server
 Takes a message and the IP's of the VM's to send the message as a slice of strings
 Messages can be encoded using golang's gobbing protocol

*/

func sendMsg(msg message, targetHosts [] string){
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(msg); err != nil {
		fmt.Println("Error :", err)
		errorCheck(err)
	}
	
	localIP, _, _ := net.ParseCIDR(currHost)
	localAddr, err := net.ResolveUDPAddr("udp", localIP.String()+ ":0")
	errorCheck(err)
	
	for _, host := range targetHosts {
		if msg.Status == "Leaving Group" || msg.Status == "Failed" {
			fmt.Println( "Propogating ", msg, " to :", host)  
		}
		ip, _, _ := net.ParseCIDR(host)
		
		serverAddr, err := net.ResolveUDPAddr("udp",ip.String()+ ":8010")
		errorCheck(err)
		
		conn, err :=net.DialUDP("udp", localAddr, serverAddr)
		errorCheck(err)
		
/*		Packet loss check and proceed
		randNum := rand.Intn(100)
		fmt.Print("Random number")
		
		*/
		_, err = conn.Write(buf.Bytes())
		errorCheck(err)
	}
	
	

}

func sendSync(){
	
}


func sendAck(){

}


//Message sent to introducer from a VM to connect to the group
func connectToIntroducer(){ 
	msg := message{currHost, "Joining", time.Now().Format(time.RFC850)}
	fmt.Println("Message transfered: ", msg)
	var targetHosts = make([]string, 1)
	targetHosts[0] = INTRODUCER

	sendMsg(msg, targetHosts)
	
}

func leaveGroup(){

}


/* Called when messages need to be propogated to the rest of the VM's. 
Example: when member leaves or fails. 
Messages are propogated to the next two members in the membershipList
If a member is not in the local membershipList then the message is ignored i.e if 
a member Ip address is to be removed and its already not present it means that its already removed and 
change has been made.
If member is there in the membershipList, call updateML to compare the timestamps and updates the membershipList
The message is then propogated to the next 2 VM's in the membershipList 
*/
func propagateMsg(msg message) {
		
}


//Called by the introducer if a new member joins the group
func sendList(){
	
}
