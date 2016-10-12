#f16-ypi1-bli12
Project: IP

Name: Yiwei Pi/ Bingbing Li

Login: ypi1 / bli12

Language: Go


================================================================================================================
                                                 DESIGN
================================================================================================================
**1. Node Interface**

    Structures:
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

	Functions:
        -- PrintInterfaces()
        -- PrintRoutes()
		-- InterfacesDown(id int)
		-- InterfacesUp(id int)
		-- PrepareAndSendPacket(cmds []string, u linklayer.UPDLink)
		-- GetRemotePhysAddr(destVirIP string) (string, int)
		-- GetLearnFrom(localVirIp string) string
          
	Description:
	    Configure a Node with a routing table and a table of its interfaces;
	    Interface down/up, and update the routing table;
	    Prepare an IP Package and send the input message to a specified virtual ip address;

	        
**2. Link Layer**

    Structures: 
	    UDPLink:
	    	socket *net.UDPConn

	Functions:
        -- InitUDP(addr string, port int) UDPLink
        -- Send(ipp ipv4.IpPackage, remoteAddr string, rePort int)
		-- Receive() ipv4.IpPackage
          
	Description:
	    Initialize a UDP connection with a given physical address;
	    Send IP Package to the specified remote IP address;
	    Receive data from socket, and convert to IP Package.


**3. IP Handler**

	Functions:
        -- HandleIpPackage(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex)
        -- CheckCsum(ipp ipv4.IpPackage) bool
		-- RunDataHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink)
		-- RunRIPHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex)
		-- ForwardIpPackage(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink, mutex *sync.RWMutex)
          
	Description:
	    When received a IP package, the IP handler first check TTL and checksums;
	    If both pass, the handler check if the IP package is locally arrived;
	    if so, check the package's protocol to determine whether pass the IP package to RunDataHandler or RunRIPHandler;
	    Otherwise, the handler forward the IP Package by looking up a next hop destination in the route table and send.
	        	    	        

**4. IP**

	Structures:
	    IpPackage:
	    	IpHeader Header
			Payload  []byte

	Functions:
        -- BuildIpPacket(payload []byte, protocol int, src string, dest string) IpPackage
        -- IpPkgToBuffer(ipp IpPackage) []byte
		-- BufferToIpPkg(buffer []byte) IpPackage
		-- Csum(header Header) int
          
	Description:
	    Build up a IP Package, based on the input payload, prtocol number, source IP address and destination IP address;
	    Convert a IP Package to []byte, to send through UDP;
	    Convert a []byte to IP Package, to send to IP Handler;
	    Calculate check sum.


**5. Threads**

    -- User input thread(main)
    -- Sending thread: keep sending out the RIP package (node's current routes) to its neighbors every 5s
    -- Receiving thread: keep receiving any data arrived through UDP connection; convert the data to IP Package and call IP Handler
    -- Timeout thread: Check the node's routes and modify expired routes to have a INFINITY cost


**6. Lock**

	-- Construct a mutex RWLock when starting a new node
    -- Every time when looking up a route in the route table: read lock/unlock (Read lock)
    -- Every time when modifying a routes in the route table: lock/unlock (Write lock)


**7. Time Out:**

    -- Every time when adding or modifying a route in the route table, initialize the route's Ttl to be current time + 12s
    -- Every 5s, the time out thread loop through all routes in current route table and check if any route's Ttl < current time:
       If so, that means the route hasn't been touched during the past 12s;
       Then modify the route's cost to INFINITY and change the route's Ttl to be current time


================================================================================================================
                                                 ROUTING ALGORITHM
================================================================================================================
Structures:

	RIP:
		Command    int    //command: 1 - request, 2 - response
		NumEntries int
		Entries    []RIPEntry
    RIPEntry:
	    Cost       int
		Address    string	

Functions:

	-- RunRIPHandler(ipPkt ipv4.IpPackage, node *pkg.Node, u linklayer.UDPLink)
	-- SendTriggerUpdates(destIpAddr string, route pkg.Entry, node *pkg.Node, u linklayer.UDPLink)

Description:

	1. When a node is on, it first send RIP request to all its neighbors;
    2. If it receives a RIP package with command=1 from its direct neighbor:
    	-- First, it will wrap all of its current routes and send back to its neighbor;
    	-- Then, it will add this direct neighbor into the route table;
    	-- Last, send trigger updates of the new added route to all its direct neighbors.
	3. If it receives a RIP package with command=2 from its direct neighbor:
		-- First, it will add the direct neighbor into the route table with cost=1, if currently there doesn't exist a route to this direct neighbor;
		-- Then, loop through all RIP entrys in the RIP package:
			---look up the node's route table to see if currently there exist a routes to the RIP entry's address:
	           ---- if the RIP entry's cost is INFINITY and the current route is learned from the RIP package's source IP address:
	                Update the route's cost to be INFINITY; 
	                Send trigger updates
	           ---- Else if the RIP entry's cost +1 < current route's cost:
	                Update the route's TTL, cost to be RIP entry's cost +1, and the route's local IP to be the RIP package's destination IP address;
	                Send trigger updates
	           ---- Else if the entry's cost+1= route's cost, and the route is learned from the RIP package's source IP address:
	                update the route's TTL  
		    ---If the RIP entry's address is not found in node's route table:
		           Add a new routes to entry's address with cost = entry's cost + 1 into the route table;
		           Send trigger updates.

Split Horizon with Poison Reverse:

    This is implemented in both the 5s periodic updates(sending thread) and trigger updates;
    Loop through all RIP entry that will be sent to a remote IP address through a route;
    Check whether the knowledge of RIP entry is learned from the route's remote IP address;
    If so, modify the RIP's cost to be INFINITY before sending.


================================================================================================================
                                               ACKNOWLEDGEMENT
================================================================================================================
I want to thank our TA Xueyang Hu for his help on designing. 
