// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

// Wraps a server-side websocket connection with its own human-readable unique
// ID. This helps to clearly map log debug and error messages to their
// respective websocket connections, thus keeping them clearly separated.
// Additionally, we also associate the capturing process (if any) with this
// connection, so we can sanely manage it.

package wsconn

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// ClosingDeadline specifies the maximum amount of time to wait for a full
// (opt. graceful) closing procedure to take.
const ClosingDeadline = 5 * time.Second

var wsupgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Things are tricky, since we have to juggle with an external process that can
// terminate or needs to be terminated, and a websocket connection that can
// error, be closed, request closing (from client) and close its side so that
// the client also closes.
//
// 1. process terminates after start (websocket open): we then need to carry out
// a graceful websocket close -- but only if the websocket is still open and not
// already closing.
//    - note to self: graceful close in progress.
//    - send close control message, informing the client about the process
//      termination reason (mutex'd with piper writer).
//    - wait for client's close control message (in websocket watcher).
//    - close websocket.
//
// 2. process fails to start (websocket open): we then need to carry out a
// graceful websocket close -- but only if the websocket is still open and not
// already closing.
//    - note to self: graceful close in progress.
//    - send close control message, informing the client about the process
//      failure reason (mutex'd with piper writer).
//    - wait for client's close control message (in websocket watcher).
//    - close websocket.
//
// 3. client closes: we then need to acknowlege the close and terminate the
// process -- please note that there's no graceful close in progress at the time
// we receive the client's close.
//    - note to self: graceful ack in progress.
//    - terminate process (if not already done so).
//    - send close control message (generic "ciao").
//    - close websocket.
//
// 4. websocket write error: as this will trigger 5. (see next) anyway and sets
// things in motion, we can just keep tucking on here, dumping any data to be
// written, but not balking either.
//
// 5. websocket read(er) error: we can only close/terminate.
//    - note to self: broken/closed.
//    - terminate process (if not already done so).
//    - close websocket.
//    - terminate reader go routine.

// WSConnState ...
type WSConnState int

const (
	// WSConnOpen declares the websocket connection being still open.
	WSConnOpen WSConnState = iota
	// WSConnClosing declares the websocket connection being in the handshake
	// for a graceful close.
	WSConnClosing
	// WSConnClosed declares the websocket connection being closed.
	WSConnClosed
)

// WSConn is a websocket connection with a unique, human-friendly ID. This
// allows differentiating multiple (concurrent) websocket connections in the
// logs.
type WSConn struct {
	*websocket.Conn             // usual (gorilla) websocket connection.
	ID              string      // unique ID string for this connection.
	mux             sync.Mutex  // protects the following fields...
	state           WSConnState // what's up???
}

// NewWSConn returns a new websocket connection wrapper that features an
// additional ID, so multiple (concurrent) websocket connections can still be
// differentiated in the logs.
func NewWSConn(w http.ResponseWriter, req *http.Request) (*WSConn, error) {
	wsconnid := petname.Generate(2, "-")
	c := &WSConn{
		ID: fmt.Sprint(wsconnid),
	}
	var err error
	c.Conn, err = wsupgrader.Upgrade(w, req, nil)
	if err != nil {
		c.Errorf("websocket upgrade process failed: %s", err.Error())
		return nil, err
	}
	return c, nil
}

// Debugf logs a formatted debug message, prefixed by the connection ID.
func (c *WSConn) Debugf(format string, args ...interface{}) {
	log.Debugf("("+c.ID+") "+format, args...)
}

// Errorf logs a formatted error message, prefixed by the connection ID.
func (c *WSConn) Errorf(format string, args ...interface{}) {
	log.Errorf("("+c.ID+") "+format, args...)
}

// Watch watches the websocket connection for any signs of closing or failure.
// Additionally, it also handles acknowledging a graceful shutdown or receiving
// a client's graceful acknowledge.
func (c *WSConn) Watch() {
	c.mux.Lock()
	state := c.state
	c.mux.Unlock()
	switch state {
	case WSConnClosed:
		return
	case WSConnClosing:
		_ = c.SetReadDeadline(time.Now().Add(ClosingDeadline))
	}
	c.Debugf("monitoring websocket connection...")
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			if cerr, ok := err.(*websocket.CloseError); ok {
				// It is not an arbitrary error, but instead a close control
				// message sent by the client. We now need to see if we need to
				// acknowledge it or if it was the final close message in the
				// handshake originally initiated by us.
				if c.state == WSConnOpen {
					// Let's try to gracefully acknowledge the close, and then
					// we're done.
					c.mux.Lock()
					c.state = WSConnClosed
					c.mux.Unlock()
					c.Debugf(
						"websocket peer started close sequence with code %d, reason \"%s\"",
						cerr.Code, cerr.Text)
					c.Debugf("acknowledging websocket close (ciao!)")
					_ = c.SetWriteDeadline(time.Now().Add(ClosingDeadline))
					_ = c.WriteMessage(
						websocket.CloseMessage,
						websocket.FormatCloseMessage(cerr.Code, "ciao"))
				} else if c.state == WSConnClosing {
					// It is already the final ack, so we're done now too.
					c.mux.Lock()
					c.state = WSConnClosed
					c.mux.Unlock()
					c.Debugf(
						"websocket peer acknowledged close with code %d, reason \"%s\"",
						cerr.Code, cerr.Text)
				}
			}
			// Any error means that the websocket is broken, and any close means
			// that we're done by now. So release resources.
			c.Debugf("websocket closed")
			c.Close()
			return
		}
		// Whatever the websocket client is sending us ... we'll ignore it. And
		// we need to keep listening in order to correctly process incomming
		// control messages.
	}
}

// InitiateGracefulClose initiates a graceful close handshake. It immediately
// returns after kicking off the close procedure. This will then cause the
// websocket reader to finish the closing handshake and finally terminating the
// capture process. If there is a problem to initiate the closing procedure,
// then the websocket will be closed immediately and the capture process
// terminated.
func (c *WSConn) InitiateGracefulClose(code int, reason string) {
	c.mux.Lock()
	state := c.state
	c.mux.Unlock()
	if state == WSConnOpen {
		c.Debugf(
			"beginning graceful websocket connection close "+
				"with code %d, reason \"%s\"...",
			code, reason)
		_ = c.SetWriteDeadline(time.Now().Add(ClosingDeadline))
		c.mux.Lock()
		c.state = WSConnClosing
		c.mux.Unlock()
		err := c.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, reason))
		if err != nil {
			c.state = WSConnClosed
			c.Errorf("sending graceful websocket close control message failed: %s", err.Error())
			c.Close()
		}
	}
}

// GracefullyClose runs a complete graceful close handshake and only returns
// after this has completed or completely failed. Use this convenience method
// when there is yet no process to also wait for or to terminate. Otherwise, use
// asynchronous InitiateGracefulClose because there's already a Watch() on this
// websocket as well as a Wait() on the capture process running in parallel.
func (c *WSConn) GracefullyClose(code int, reason string) {
	c.mux.Lock()
	state := c.state
	c.mux.Unlock()
	if state == WSConnOpen {
		c.InitiateGracefulClose(code, reason)
		c.Watch()
	}
}
