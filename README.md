This is actually the draft of the report. Readme will be added shortly

Protocol : SWIM Protocol

Language : GO

Topology for Group Membership: Extended Ring Structure


Assumptions
-----------
Minimum number of nodes in the group is 4

A maximum of 3 simultaneous node failures can happen in the group

Design
------
Each machine maintains a sorted, full membership list which is used to monitor the next three
machines in the system ring by using SYN->ACK messages. The protocol uses an
introducer to accept and handle pings from newly joining members. Every process stores the IP
address of the introducer as a constant so that at startup it knows what VM to ping when
joining the group. 

Every process initiates 2 servers at startup: one for accepting incoming messages and one for accepting an 
updated membership sent by the introducer when a newmember joins.
Membership lists are sent over port 8011 and messages are propagated over port 8010.
UDP connections are used to send messages and data to and from machines within the group.

Machines in the group are stored in the membership list as member structs containing the machines IP
address and the timestamp for when it was added to the system. The membership list is always
sorted on the basis of the IP address of the system and each system takes care of the (n+1)%N, (n+2)%N
and (n+3)%N members on its list, where N is the size of the membership list.
We use a message structure which contains three fields: (1) Host which sent the message, (2) Message
info/status, and (3) Timestamp for when the time the message was sent. The possible messages
are explained below.

Joining : A node can join by sending a message to the introducer with a status of “Joining”. The
introducer then updates its own membership list, and propagates this updated list to
everybody on its membership list. This way, each machine updates its membership list and can
appropriately change the members it is monitoring.

Leaving: When leaving voluntarily, the node informs the previous 3 nodes that it is leaving. The 2 nodes 
now propogate the message throughout the ring, and the membership list in every other node gets updated.

Crashing: A crashing or failure happens when the node does not sent a heartbeat back to the monitoring nodes(ACK)
When this happens, node is assumed to be down and the failure message is propogated throughout the ring to make sure the 
other nodes remove this node from their membership list. If the node has gone down and once treated as failed node, 
it will not be part of the group if it comes back up. It has to send a new join request to be part of the group again.
The time in which the node has to send an ACK to show that it's alive is 250 milliseconds.

IsAlive: This message is sent by the introducer to all the nodes in the local membership file when the introducer comes back up after failure or shutdown. This is useful because other nodes may have left the group while the introduer was down, and hence introducer needs to make changes to its membership list accordingly. Also, new nodes would not have joined because the introducer was down. This setup preserves the consistency of the membership list in the case of introducer failures.

IamAlive: This message is sent as a response by all the nodes still in the group for the isAlive message sent by the introducer. It gives an acknowledgement to the introducer that it is still part of the group.

Any node in the group can do a distributed Grep of the logFile to find out about the events that have been logged,
and thus will get an idea of how the state has been reached. Distributed grep will be run on current members of the group 
as is available from the membership list


State 1
------
Introducer starts and Checks for MList: -> shows only Introducer in the MList 

State 2
------
If introducer tries to join the group: -> shows message "I am the introducer" to indicate that introducer should always be part of the group 

State 3
------
Non introducer node checks Mlist without joining: -> shows not allowed since only group members can have access to the mList 

State 4
------
Non introducer leaves without joining: -> shows not part of the group to leave the group

State 5
------
Non introducer joins the group: -> shows "joined the group" and updates the Mlist of introducer and finally propagates to all the nodes. Change will be visible in all the nodes eventually including the node which just joined.

State 6	
------
Any node leaves the group voluntarily: -> shows propagating Leaving message to next 3 VM's and Mlist gets updated everywhere -> done

State 7
------
Up to 3 nodes fail: -> shows propagating Failed message to next 3 VM's and Mlist gets updated everywhere

State 8
------
Introducer goes down: -> Joining does not happen, but all other functions intact and will work seamlessly

State 9
------
Introducer comes back up and does a fresh start with new Group and no other group members.

State 10
-------
Introducer comes back up and wants to use the existing Mlist (the Mlist before it went down) : -> Introducer takes the Mlist locally and sends isAlive messages to the group members to see if they are still up and on the basis of the iamAlive results, introducer updates the Mlist. State 9 or 10 can be decided based on user input.

Notes
-----
1. When the membership list gets updated, the timers to keep track of the heartbeat will be reset.
2. When the introducer is down, new node joining requests cannot be processed until the introducer comes back up.
3. However, when the introducer is down, other functions like leaving and failure detections will be active
4. Grep functionality is taken from https://github.com/vimalphilip/DistributedGrep and made available as a an option





