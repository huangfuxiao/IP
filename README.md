#f16-ypi1-bli12
Project: TCP Over IP

Name: Yiwei Pi/ Bingbing Li

Login: ypi1 / bli12

Language: Go


================================================================================================================
                                                 TCP DESIGN
================================================================================================================
## 1. TCP Connection(TCB):

#### Structures:
	    TCB:
	    	LocalAddr 		string
	    	Fd        		int
			State           tcp.State
			Addr            SockAddr
	        Seq             int
	        Ack             int
	        RecvW           RecvWindow
	        SendW           SendWindow
	        node            *pkg.Node
	        u               linklayer.UDPLink
	        PIFCheck        map[int]*PkgInFlight
	        Check           map[int]bool
	        BlockWrite      bool
	        BlockRead       bool
	        ShouldClose     bool
	    	
	    PkgInFlight:
	    	Length          int
			Count           int
	        Ipp             ipv4.IpPackage
	        Addr            string
	        Port            int
	
		SockAddr:
			LocalAddr       string
			LocalPort       int
			RemoteAddr      string
			RemotePort      int

#### Functions:
* BuildTCB(fd int, node *pkg.Node, u linklayer.UDPLink) TCB
* SendCtrlMsg(ctrl int, c bool, notestb bool, ws int)
* SendDataThread()
* SendData(payload []byte, ws int)
* To4byte(addr string) [4]byte
* CheckACK(idx int, tcb *TCB, count int, ipp ipv4.IpPackage, addr string, port int)
* DataACKThread()
          
#### Description:
* Configure a TCP connection with:
* 	Basic information: Socket address(src IP, dst IP, src port, dst port), node, UPD link    
* 	State machine: We have implemented most TCP state machine
* 	Send window: We implement the send window by a sending buffer and couple pointers: LastByteWritten, LastByteAcked, LastByteSent, BytesInFlight;
	             Sending buffer stores data that need to be sent;
	             BytesInFlight stores the number of data bytes that has already been sent before and waits for ack;
	             When the user needs to send data, these data may be appended into the tail of the sender buffer;
	             When the sending data thread is ready to send data, the thread will pop data from the head of the writing buffer and increase the flight bytes number;
	             When acks arrives, the flight bytes number will be decreased. 
* 	Receive window: We implement receive window by a receiving buffer and couple pointers: LastSeq, LastByteRead, NextByteExpected;
	                Received data bytes are placed in the receiving buffer based on their sequence number;
	                If sequence number is not inside the sliding window, it will be dropped;
	                If data is read out, LastByteRead would be advanced;
	                If data is written continuous to the receiving buffer, NextByteExpected would be advanced till the end of continuous written data.
* 	Sequence/acknowledge number: Sequence number will be generated randomly when the client initiates a connection. 
	                             Also, during the three-way handshake, these two TCP peers will synchronize ack numbers. 
* Flow control:   The sender can know the advertised window size (available buffer size) of receiver by checking the window size field in the TCP head. 
                When this field becomes 0, the sender will keep sending 1-byte segments to probe the remote window size.	    
* Timeout:	    
Establishing or teardowning a connection timeout: it will do 3 re-transmit SYN or FIN for establishing or teardowning a new connection.
	            Once it receives a valid ack back, it will cancel the timeout. 
	            Data sending timeout: When the sender fails to receive valid ACK, it will retansmit all flight date for at most 5 times.
* Lossy:

	        
## 2. Socket API:

#### Structures: 
	    SocketManager:
	    	Fdnum           int
			Portnum         int
			FdToSocket      map[int]*TCB
			AddrToSocket    map[SockAddr]*TCB
			Interfaces      map[string]bool

#### Functions:
* BuildSocketManager(interfaceArray []*pkg.Interface) SocketManager
* PrintSockets()
* V_socket(node *pkg.Node, u linklayer.UDPLink) int
* V_bind(socket int, addr string, port int) int
* V_listen(socket int) int
* V_connect(socket int, addr string, port int) int
* V_accept(socket int, addr string, port int, node *pkg.Node, u linklayer.UDPLink) int
* V_read(socket int, nbyte int, check string) (int, []byte)
* V_write(socket int, data []byte) int
* V_shutdown(socket int, ntype int) int 
* V_close(socket int) int
* CloseThread() 
          
#### Description:
Configure a socket manager with:    
	Basic information: socket number, port number, socket number to TCB map, socket address to TCB map, interfaces list
Implement API functions/methods to TCP implementation;    
Loop through all TCP connections every 3 seconds; If the connection is closed, remove it from maps.


## 3. TCP Handler:

#### Description:
When received a IP package, and the protocol=6, the package would be passed to TCP handler;    
First check the TCP checksum; If the packet fail to pass the checksum, drop the packet;
Find the corresponding connection (transmission control block - TCB) and call receive function for that TCB;
Split cases when payload is empty or not. 
When the payload is empty, the tcp packet is a control flag packet.
When the payload is not empty, the tcp packet is an actual data packet.

   

## 4. Threads:

* User input thread(main);
* TCP Handler thread: keep receiving TCP packets, and pass to corresponding TCB;
* Data sending thread: keep sending data if there is some data in sending buffer, which is created after the three-way handshake finished;
* Receive file thread: application level, only when starting to receive data into file;
* Sending file thread: application level, only when starting to send data from file;
* Timeout thread: Establishing or teardowning a connection timeout, and Data sending timeout.


## 5. Lock:

* Each sliding window will have a mutex lock.
* Sending/receiving buffer: the mutex lock controls each buffer, which aims to avoid reading and writing at the same time.


================================================================================================================
                                                PERFORMANCE
================================================================================================================    




================================================================================================================
                                                    BUGS
================================================================================================================




================================================================================================================
                                                IP DESIGN
================================================================================================================
## 1. Node Interface:

#### Structures:
	    Node:
	    	LocalAddr 		string
	    	Port      		int
	    	InterfaceArray  []*Interface 
	    	RouteTable 		map[string]Entry
	    Interface:
			Status    		int
			Src        		string
			Dest       		string
			RemotePort 		int
			RemoteAddr 		string
		Entry:
			Dest 			string
			Next 			string
			Cost 			int
			Ttl  			int64	

#### Functions:
* PrintInterfaces()
* PrintRoutes()
* InterfacesDown(id int)
* InterfacesUp(id int)
* PrepareAndSendPacket(cmds []string, u linklayer.UPDLink)
* GetRemotePhysAddr(destVirIP string) (string, int)
* GetLearnFrom(localVirIp string) string
          
#### Description:
Configure a Node with a routing table and a table of its interfaces;    
Interface down/up, and update the routing table;    
Prepare an IP Package and send the input message to a specified virtual ip address;

	        
## 2. Link Layer:

#### Structures: 
	    UDPLink:
	    	socket *net.UDPConn

#### Functions:
* InitUDP(addr string, port int) UDPLink
* Send(ipp ipv4.IpPackage, remoteAddr string, rePort int)
* Receive() ipv4.IpPackage
          
#### Description:
Initialize a UDP connection with a given physical address;    
Send IP Package to the specified remote IP address;    
Receive data from socket, and convert to IP Package.


## 3. IP Handler:

#### Functions:
* HandleIpPackage(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex, manager *api.SocketManager)
* CheckCsum(ipp ipv4.IpPackage) bool
* RunDataHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink)
* RunRIPHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex)
* RunTCPHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex, manager *api.SocketManager)
* ForwardIpPackage(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex)
          
#### Description:
When received a IP package, the IP handler first check TTL and checksums;    
If both pass, the handler check if the IP package is locally arrived;    
if so, check the package's protocol to determine whether pass the IP package to RunDataHandler or RunRIPHandler or RunTCPHandler;    
Otherwise, the handler forward the IP Package by looking up a next hop destination in the route table and send.    
	        	    	        

## 4. IP:

#### Structures:
	    IpPackage:
	    	IpHeader Header
			Payload  []byte

#### Functions:
* BuildIpPacket(payload []byte, protocol int, src string, dest string) IpPackage
* IpPkgToBuffer(ipp IpPackage) []byte
* BufferToIpPkg(buffer []byte) IpPackage
* Csum(header Header) int
          
#### Description:
Build up a IP Package, based on the input payload, prtocol number, source IP address and destination IP address;    
Convert a IP Package to []byte, to send through UDP;    
Convert a []byte to IP Package, to send to IP Handler;    
Calculate check sum.    


## 5. Threads:

* User input thread(main)
* Sending thread: keep sending out the RIP package (node's current routes) to its neighbors every 5s
* Receiving thread: keep receiving any data arrived through UDP connection; convert the data to IP Package and call IP Handler
* Timeout thread: Check the node's routes and modify expired routes to have a INFINITY cost


## 6. Lock:

Construct a mutex RWLock when starting a new node    
Every time when looking up a route in the route table: read lock/unlock (Read lock)    
Every time when modifying a routes in the route table: lock/unlock (Write lock)    


## 7. Time Out:

Every time when adding or modifying a route in the route table, initialize the route's Ttl to be current time + 12s    
Every 5s, the time out thread loop through all routes in current route table and check if any route's Ttl < current time:    
If so, that means the route hasn't been touched during the past 12s;    
Then modify the route's cost to INFINITY and change the route's Ttl to be current time    


================================================================================================================
                                                 ROUTING ALGORITHM
================================================================================================================
#### Structures:

	RIP:
		Command    int    //command: 1 - request, 2 - response
		NumEntries int
		Entries    []RIPEntry
    RIPEntry:
	    Cost       int
		Address    string	

#### Functions:

* RunRIPHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink)
* SendTriggerUpdates(destIpAddr string, route pkg.Entry, node *pkg.Node, u linklayer.UDPLink)

#### Description:
When a node is on, it first send RIP request to all its neighbors;    
If it receives a RIP package with command=1 from its direct neighbor: 

* First, it will wrap all of its current routes and send back to its neighbor;
* Then, it will add this direct neighbor into the route table;
* Last, send trigger updates of the new added route to all its direct neighbors.

If it receives a RIP package with command=2 from its direct neighbor:    

* First, it will add the direct neighbor into the route table with cost=1, if currently there doesn't exist a route to this direct neighbor;
* Then, loop through all RIP entrys in the RIP package:
* look up the node's route table to see if currently there exist a routes to the RIP entry's address:
* if the RIP entry's cost is INFINITY and the current route is learned from the RIP package's source IP address:
  * Update the route's cost to be INFINITY; 
  * Send trigger updates
* Else if the RIP entry's cost +1 < current route's cost:
  * Update the route's TTL, cost to be RIP entry's cost +1, and the route's local IP to be the RIP package's destination IP address;
  * Send trigger updates
* Else if the entry's cost+1= route's cost, and the route is learned from the RIP package's source IP address:
  * update the route's TTL  
* If the RIP entry's address is not found in node's route table:
  * Add a new routes to entry's address with cost = entry's cost + 1 into the route table;
  * Send trigger updates.

#### Split Horizon with Poison Reverse:

This is implemented in both the 5s periodic updates(sending thread) and trigger updates;    
Loop through all RIP entry that will be sent to a remote IP address through a route;    
Check whether the knowledge of RIP entry is learned from the route's remote IP address;    
If so, modify the RIP's cost to be INFINITY before sending.    


================================================================================================================
                                               ACKNOWLEDGEMENT
================================================================================================================
I want to thank our TA Xueyang Hu for his help on designing. 
