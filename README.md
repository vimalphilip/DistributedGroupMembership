
Protocol : SWIM Protocol

Language : GO


Assumptions: 

Minimum number of nodes in the group is 4

A maximum of 3 simultaneous node failures can happen in the group

Design
------

Each machine maintains a sorted, full membership list which is used to monitor the next two
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

Leaving: When leaving voluntarily, the node informs the previous 2 nodes that it is leaving. The 2 nodes 
now propogate the message throughout the ring, and the membership list in every other node gets updated.

Crashing: A crashing or failure happens when the node does not sent a heartbeat back to the monitoring nodes(ACK)
When this happens, node is assumed to be down and the failure message is propogated throughout the ring to make sure the 
other nodes remove this node from their membership list. If the node has gone down and once treated as failed node, 
it will not be part of the group if it comes back up. It has to send a new join request to be part of the group again.
The time in which the node has to send an ACK to show that it's alive is 250 milliseconds.

Notes
-----
1. When the membership list gets updated, the timers to keep track of the heartbeat will be reset.
2. When the introducer is down, new node joining requests cannot be processed until the introducer comes back up
3. However, when the introducer is down, other functions like leaving and failure detections will be active





