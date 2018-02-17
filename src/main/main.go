package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
	"os"
	"net"
	"sort"
	"sync"
)

//Declare constants
const INTRODUCER = "172.31.36.139/20"		//IP Address of the introducer 
const FILE_PATH = "MembershipList.txt"		//File path of membership list
const MAX_TIME = time.Millisecond * 2500	//Max time a VM has to wait for the Syn/Ack message
const MIN_HOSTS = 5 						//Minimum number of VM's in the group before Syn/Ack begins


var currHost string    						//	IP of the local machine
var isConnected int 						//  1(Connected) or 0(Not connected) -> Boolean value to check if machine is currently connected to the group
var membershipList = make([]member, 0)		//Contains all members connected to the group


//We need 2 timers to keep track of the connected nodes
var timers [2]*time.Timer
//We also need 2 flags to keep track of nodes voluntarily leaving or crashed
//1 = timers were forcefully stopped
var resetFlags [2]int

//Mutex used for membershipList and timers
var mutex = &sync.Mutex{}

//Message object passed b/w client and server
type message struct {
	Host string
	Status string
	Timestamp string
}

//Information kept for each VM in the group stored in membership list
type member struct {
	Host string
	Timestamp string
}

//type and functions used to sort membershipLists
type memList []member
//TODO functions to sort the membershipList
func (slice memList) Len() int           { return len(slice) }
func (slice memList) Less(i, j int) bool { return slice[i].Host < slice[j].Host }
func (slice memList) Swap(i, j int)      { slice[i], slice[j] = slice[j], slice[i] }

	

//Log files for error and info logging
var logfile *os.File
var errlog *log.Logger
var infolog *log.Logger

//For simulating packet loss in percent
const PACKET_LOSS = 0

var packets_lost int

func main(){
	
	//setup and initialize starting variables
	setupAndInitialize()
	//start all the VM's to receive connections and messages from 1) Connected VM's 2) Introducer when new VM's join 
	
	//1)
	go messageServer()
	//2)
	go introducerMachineServer()


	//Reader to take console input from the user
	reader := bufio.NewReader(os.Stdin)
	
	//Check if the VM is the introducer
	if currHost == INTRODUCER {
		//If membershipList file exists, check if user wants to restart server using
		//the file or start a new group
		//If VM is the introducer, follow protocol for storing membershipList as a local file
		if _, err := os.Stat(FILE_PATH); os.IsNotExist(err) {
			writeMLtoFile()
		} else {
			//TODO Logic to read file from the directory and convert to membershipList
			}
	}
	
	
	// start sending sync functions and check for acks in seperate threads
	go sendSyn()
	go checkLastAck(1)
	go checkLastAck(2)
	
	
	//Take inputs from the console on what to do?
	//4 options 
	// 1. Print the membershiplist
	// 2. Show this machine's IP address
	// 3. Join the group
	// 4. Leave the group
	
	for {
		fmt.Println("1 -> Print membership list")
		fmt.Println("2 -> Show IP address of this host")
		fmt.Println("3 -> Join group")
		fmt.Println("4 -> Leave group\n")
		input, _ := reader.ReadString('\n')
		switch input {
		case "1\n":
			for _, element := range membershipList {
				fmt.Println(element)
			}
		case "2\n":
			fmt.Println(currHost)
		case "3\n":
			if currHost != INTRODUCER  {
				if isConnected == 0 {
					fmt.Println("Joining group")
					connectToIntroducer()
					infoCheck(currHost + " is connecting to introducer")
					isConnected = 1
				} else {
					fmt.Println("I am already connected to the group")
				}
			} else {
				fmt.Println("I am the introducer")
			}
		case "4\n":
			if isConnected == 1 {
				fmt.Println("Leaving group")
				leaveGroup()
				infoCheck(currHost + " left group")
				os.Exit(0)

			} else {
				fmt.Println("You are currently not connected to a group")
			}
		default:
			fmt.Println("Invalid command")
		}
		fmt.Println("\n\n")
	}
}
	
	


func messageServer(){
	// We need to implement UDP to make it light weight for the heartbeat messages
	serverAddress, err := net.ResolveUDPAddr("udp", ":8010")
	errorCheck(err)
	serverConn, err := net.ListenUDP("udp", serverAddress)
	errorCheck(err)
	
	defer serverConn.Close()
	
	buf := make([]byte, 1024)
	
	for {
			// Constantly Listening
			msg := message{}
			n, _, err := serverConn.ReadFromUDP(buf)
			errorCheck(err)
			gob.NewDecoder(bytes.NewReader(buf[:n])).Decode(&msg)
			
			//Different cases 
			switch msg.Status {
				case "Joining":
						node := member{msg.Host, time.Now().Format(time.RFC850)}  //TODO check time format
						//todo check all conditions and if ok append to the membership list
						if checkTimeStamp(node) == 0 {
							mutex.Lock()
							resetTimers()
							membershipList = append(membershipList, node)
							sort.Sort(memList(membershipList))
							mutex.Unlock()							
						}					
						//check the possible conditions and if all good write to File. Also check error conditions
						go writeMLtoFile();
						sendList()
				case "Leaving":
						mutex.Lock()
						resetTimers()
						propagateMsg(msg)
						mutex.Unlock()	
				case "SYN":
						infoCheck("Syn received from: "+msg.Host)
						sendAck(msg.Host)
				/*	if ack, check if ip that sent the message is either (currIndex + 1)%N or (currIndex + 2)%N
					and reset the corresponding timer to MAX_TIME*/
				case "ACK":
						if msg.Host == membershipList[(getIndex()+1)%len(membershipList)].Host {
							infoCheck("ACK received from "+msg.Host)
							timers[0].Reset(MAX_TIME)
						} else if msg.Host == membershipList[(getIndex()+2)%len(membershipList)].Host {
							infoCheck("ACK received from "+msg.Host)
							timers[1].Reset(MAX_TIME)
						}
						
						//if message status is failed, propagate the message (timers will be taken care of in checkLastAck
				case "Failed":
						//resetTimers taken care in checkLastAck
						mutex.Lock()
						propagateMsg(msg)  //Ideally the logic executed should be same as leaving */
						mutex.Unlock()		
			}
	}
	
	
}

func introducerMachineServer() {
	//Listens to messages from introducer
	
	serverAddress, err := net.ResolveUDPAddr("udp", ":8011")
	errorCheck(err)
	serverConn, err := net.ListenUDP("udp", serverAddress)
	errorCheck(err)
	
	defer serverConn.Close()
	
	buf := make([]byte, 1024)
	
	for {
		// Constantly Listening 
		mList := make([]member, 0)
		n, _, err := serverConn.ReadFromUDP(buf)
		err = gob.NewDecoder(bytes.NewReader(buf[:n])).Decode(&mList)
		errorCheck(err)

		//restart timers if membershipList is updated
		mutex.Lock()	
		resetTimers()
		membershipList = mList
		mutex.Unlock()

		var msg = "New VM joined the group: \n\t["
		var size = len(mList) - 1
		for i, host := range mList {
			msg += "(" + host.Host + " | " + host.Timestamp + ")"
			if i != size {
				msg += ", \n\t"
			} else {
				msg += "]"
			}
		}
		infoCheck(msg)
	}
}

//VM's are marked as failed if they have not responded with an ACK within MAX_TIME
//2 checkLastAck calls persist at any given time, one to check the VM at (currIndex + 1)%N and one to
//check the VM (currIndex + 2)%N, where N is the size of the membershipList
//relativeIndex can be 1 or 2 and indicates what VM the function to watch
//A timer for each of the two VM counts down from MAX_TIME and is reset whenever an ACK is received (handled in
// messageServer function.
//Timers are reset whenever the membershipList is modified
//The timer will reach 0 if an ACK isn't received from the corresponding VM
// within MAX_TIME, or the timer is reset. If a timer was reset, the corresponding resetFlag will be 1
// and indicate that checkLastAck should be called again and that the failure detection should not be called
//If a timer reaches 0 because an ACK was not received in time, the VM is marked as failed and the message is
//propagated to the next 2 VM's in the membershipList. Both timers are then restarted.
func checkLastAck(relativeIndex int) {
	//Wait until number of members in group is at least MIN_HOSTS before checking for ACKs
	for len(membershipList) < MIN_HOSTS {
		time.Sleep(100 * time.Millisecond)
	}

	//Get host at (currIndex + relativeIndex)%N
	host := membershipList[(getIndex()+relativeIndex)%len(membershipList)].Host
	infoCheck("Checking "+string(relativeIndex)+ ": "+host )
	

	//Create a new timer and hold until timer reaches 0 or is reset
	timers[relativeIndex-1] = time.NewTimer(MAX_TIME)
	<-timers[relativeIndex-1].C

	/*	3 conditions will prevent failure detection from going off
		1. Number of members is less than the MIN_HOSTS
		2. The target host's relative index is no longer the same as when the checkLastAck function was called. Meaning
		the membershipList has been updated and the checkLastAck should update it's host
		3. resetFlags for the corresponding timer is set to 1, again meaning that the membership list was updated and
		checkLastack needs to reset the VM it is monitoring.*/
	mutex.Lock()
	if len(membershipList) >= MIN_HOSTS && getRelativeIndex(host) == relativeIndex && resetFlags[relativeIndex-1] != 1 {
		msg := message{membershipList[(getIndex()+relativeIndex)%len(membershipList)].Host, "Failed", time.Now().Format(time.RFC850)}
		fmt.Print("Failure detected: ")
		fmt.Println(msg.Host)
		propagateMsg(msg)

	}
	//If a failure is detected for one timer, reset the other as well.
	if resetFlags[relativeIndex-1] == 0 {
		infoCheck("Force stopping timer "+string(relativeIndex))
		resetFlags[relativeIndex%2] = 1
		timers[relativeIndex%2].Reset(0)
	} else {
		resetFlags[relativeIndex-1] = 0
	}

	mutex.Unlock()
	go checkLastAck(relativeIndex)

}	

