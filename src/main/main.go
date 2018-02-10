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
)

//Declare constants
const INTRODUCER = "<IPAddress>"			//1.	IP Address of the introducer 
const FILE_PATH = "MembershipList.txt"		//2.  	File path of membership list
const MAX_TIME = time.Millisecond * 2500	//3.	Max time a VM has to wait for the Syn/Ack message


var currHost string    				//	IP of the local machine
var isConnected int 				//  1(Connected) or 0(Not connected) -> Boolean value to check if machine is currently connected to the group
var membershipList = make([]member, 0)	//Contains all members connected to the group


//We need 2 timers to keep track of the connected nodes
var timers [2]*time.Timer
//We also need 2 flags to keep track of nodes voluntarily leaving or crashed
//1 = timers were forcefully stopped
var resetFlags [2]int


//Message object passed b/w client and server
type message struct{
	host string
	status string
	timestamp string
}

//Information kept for each VM in the group stored in membership list
type member struct{
	host string
	timestamp string
}

//type and functions used to sort membershipLists
type memList []member
//TODO functions to sort the membershipList
func (slice memList) Len() int           { return len(slice) }
func (slice memList) Less(i, j int) bool { return slice[i].host < slice[j].host }
func (slice memList) Swap(i, j int)      { slice[i], slice[j] = slice[j], slice[i] }

	

//Log files for error and info logging
var logfile *os.File
var errlog *log.Logger
var infolog *log.Logger

func main(){
	
	//start
	//start all the VM's to receive connections and messages from 1) Connected VM's 2) Introducer when new VM's join 
	
	//1)
	go messageServer()
	//2)
	go introducerMachineServer()


	//Reader to take console input from the user
	reader := bufio.NewReader(os.Stdin)
	
	//  check if the current host = introducer and if yes follow the steps to store membershipList as a local file
	   // if membership file exists, check if user wants to restart the server using the file or start a new group
	   
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
			if currHost != INTRODUCER && isConnected == 0 {
				fmt.Println("Joining group")
				connectToIntroducer()
				infoCheck(currHost + " is connecting to introducer")
				isConnected = 1
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
	serverAddress, err := net.ResolveUDPAddr("udp", "address")
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
			switch msg.status {
				case "Joining":
						node := member{msg.host, time.Now().Format(time.RFC850)}  //TODO check time format
						//todo check all conditions and if ok append to the membership list
						if checkTimeStamp(node) == 0 {
							//create a lock
							resetTimers()
							membershipList = append(membershipList, node)
							sort.Sort(memList(membershipList))
							//release the lock
						}					
						//check the possible conditions and if all good write to File. Also check error conditions
						go writeMLtoFile();
						sendList()
				case "Leaving":
						propagateMsg(msg)
				case "Acknowledgement":
						if msg.host == membershipList[(getIndex()+1)%len(membershipList)].host {
							fmt.Print("ACK received from ")
							fmt.Println(msg.host)
							timers[0].Reset(MAX_TIME)
						}else if msg.host == membershipList[(getIndex()+2)%len(membershipList)].host {
							fmt.Print("ACK received from ")
							fmt.Println(msg.host)
							timers[1].Reset(MAX_TIME)
						}
			}
	}
	
	
}

func introducerMachineServer() {
	//Listens to messages from introducer
	
	serverAddress, err := net.ResolveUDPAddr("udp", "address")
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
		//Make a lock
		resetTimers()
		membershipList = mList
		//Release a lock

		var msg = "New VM joined the group: \n\t["
		var size = len(mList) - 1
		for i, host := range mList {
			msg += "(" + host.host + " | " + host.timestamp + ")"
			if i != size {
				msg += ", \n\t"
			} else {
				msg += "]"
			}
		}
		infoCheck(msg)
	}
}

func sendSyn(){
	
}

func checkLastAck(index int){
	
}
