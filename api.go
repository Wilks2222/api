package api

/* This API receive request to get status of Probe and return the value.

   Request are done over TCP by sending the ID of the Probe follow by line feed /n.
   the API return the following values:

	100 - config.StatusUP for OK

   	15 - config.StatusMAINTENANCE for WARN

   	25 - config.StatusDEGRADATION for ALARM

   	75 - config.StatusJEOPARDY for CRITICAL

   	0 - config.StatusDOWN for DOWN

*/

import (
	"bufio"
	"github.com/antigloss/go/logger"
	"github.com/Wilks2222/config"
	"github.com/Wilks2222/database"
	"net"
	"os"
	"strconv"
	"time"
)

/* Channel to handle the connection in a thread
 */

var connectionsChannel chan net.Conn

/* Start the TCP API, listen for connection and API request.
 */

func Start(listenIP string, HandleConnectionThreads int) {

	connectionsChannel = make(chan net.Conn)

	// make sure there is at least one thread that start.
	if HandleConnectionThreads <= 0 {
		HandleConnectionThreads = 1
	}

	// Listen for incoming connections.
	l, err := net.Listen("tcp", listenIP)
	if err != nil {
		logger.Panic("Unable to start TCP-API: " + err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()

	// start the threads to handle the TCP connections.
	for i := 0; i < HandleConnectionThreads; i++ {
		go handleRequest()
	}

	// Now Listen for TCP requests coming in on the socket.
	for {
		// Listen for incoming connections.
		conn, err := l.Accept()
		if err != nil {
			logger.Error("Error accepting: " + err.Error())
			continue
		}
		// Handle connections in a new goroutine.
		connectionsChannel <- conn
	}

}

// Thread tHAT Handles incoming requests, default setup start 64 of those threads

func handleRequest() {

	for conn := range connectionsChannel {

		// each iteration is a connection that open, now listen for incoming
		// request.  Request are an ID send follow by a newline code.

		timeoutDuration := 5 * time.Second
		bufReader := bufio.NewReader(conn)

		// Set a deadline for reading. Read operation will fail if no data
		// is received after deadline.
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))

		// Read tokens delimited by newline
		id, err := bufReader.ReadBytes('\n')
		if err != nil {
			//logger..Trace("Error reading:" + err.Error())
			continue
		}

		logger.Trace("receive request TCP-API for ID (" + string(id) + ")")

		if len(id) > 0 && string(id) == "OWLSO-SERVER" {

			conn.Write([]byte(strconv.Itoa(config.StatusUP)))

		} else {
			status, err := database.GetStatus(id)
			if err == nil {

				// we found the ID status, now convert the status value
				// into text to be return.

				conn.Write([]byte(strconv.Itoa(status)))

			}
		}

		// Send a response back to person contacting us.
		// Close the connection when you're done with it.
		conn.Close()
	}
}
