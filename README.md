
Protocol : SWIM Protocol
Language : GO

Design
------
Each machine maintains a sorted, full membership list which is used to monitor the next two
machines in the system ring by using SYN->ACK->SYN messages. The protocol uses an
introducer to accept and handle pings from newly joining members. Every process stores the IP
address of the introducer as a constant so that at startup it knows what VM to ping when
joining the group. 

Every process initiates 2 servers at startup: one for accepting incoming
messages and one for accepting an updated membership sent by the introducer when a new
member joins.
Membership lists are sent over port <portnumber> and messages are propagated over
port <portnumber>. UDP connections are used to send messages and data to and from machines within
the group.
Machines in the group are stored in the membership list as member structs containing the machines IP
address and the timestamp for when it was added to the system. The membership list is always
sorted on the basis of the IP address of the system and each system takes care of the (n+1)%N, (n+2)%N
and (n+3)%N members on its list, where N is the size of the membership list. We use a message
structure which contains three fields: (1) Host which sent the message, (2) Message
info/status, and (3) Timestamp for when the time the message was sent. The possible messages
are explained below.

Joining : A node can join by sending a message to the introducer with a status of “joining”. The
introducer then updates its own membership list, and propagates this updated list to
everybody on its membership list. This way, each machine updates its membership list and can
appropriately change the members it is monitoring.
Leaving: 
Crashing: 
