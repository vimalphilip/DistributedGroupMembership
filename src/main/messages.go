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
		errorCheck(err)
	}
	
	localIP, _, _ := net.ParseCIDR(currHost)
	localAddr, err := net.ResolveUDPAddr("udp", localIP.String()+ ":0")
	errorCheck(err)
	
	for _, host := range targetHosts {
		if msg.Status == "Leaving" || msg.Status == "Failed" {
			fmt.Println( "Propogating ", msg, " to :", host)  
		}
		ip, _, _ := net.ParseCIDR(host)
		
		serverAddr, err := net.ResolveUDPAddr("udp",ip.String()+ ":8010")
		errorCheck(err)
		fmt.Println("Local Address: ", localAddr, "ServerAddress: ", serverAddr)
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

//Leave group
//Message sent to previous 2 VM's in membershiplist notifying that the VM is leaving the group
func leaveGroup(){
	msg := message{currHost, "Leaving", time.Now().Format(time.RFC850)}
	
	var targetHosts = make([]string, 2) //make a string array of size 2 to send to previous 2 VM's
	for i := 1; i < 3; i++ {
		var targetHostIndex = (getIndex() - i) % len(membershipList)
		if targetHostIndex < 0 {
			targetHostIndex = len(membershipList) + targetHostIndex
		}
		targetHosts[i-1] = membershipList[targetHostIndex].Host
	}

	sendMsg(msg, targetHosts)

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
	var hostIndex = -1
	for i, element := range membershipList {
		if msg.Host == element.Host {
			hostIndex = i
			break
		}
	}
	if hostIndex == -1 {  // case where node is already removed and hence could not find the node in the Mlist
		return
	}

	// msgCheck(msg)
	updateML(hostIndex, msg)

	var targetHosts = make([]string, 2)
	targetHosts[0] = membershipList[(getIndex()+1)%len(membershipList)].Host
	targetHosts[1] = membershipList[(getIndex()+2)%len(membershipList)].Host

	sendMsg(msg, targetHosts)
		
}


//Called by the introducer if a new member joins the group
func sendList() {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(membershipList); err != nil {
		errorCheck(err)
	}
	for index, element := range membershipList {
		if element.Host != currHost {
			ip, _, _ := net.ParseCIDR(membershipList[index].Host)

			ServerAddr, err := net.ResolveUDPAddr("udp", ip.String()+":8011")
			errorCheck(err)

			localip, _, _ := net.ParseCIDR(currHost)
			LocalAddr, err := net.ResolveUDPAddr("udp", localip.String()+":0")
			errorCheck(err)

			conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
			errorCheck(err)

			_, err = conn.Write(buf.Bytes())
			errorCheck(err)
		}
	}
}
