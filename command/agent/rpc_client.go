package agent

import (
	"bufio"
	"fmt"
	"github.com/ugorji/go/codec"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

var (
	clientClosed = fmt.Errorf("client closed")
)

type seqCallback struct {
	handler func(*responseHeader)
}

func (sc *seqCallback) Handle(resp *responseHeader) {
	sc.handler(resp)
}
func (sc *seqCallback) Cleanup() {}

// seqHandler interface is used to handle responses
type seqHandler interface {
	Handle(*responseHeader)
	Cleanup()
}

// RPCClient is the RPC client to make requests to the agent RPC.
type RPCClient struct {
	seq uint64

	conn      *net.TCPConn
	reader    *bufio.Reader
	writer    *bufio.Writer
	dec       *codec.Decoder
	enc       *codec.Encoder
	writeLock sync.Mutex

	dispatch     map[uint64]seqHandler
	dispatchLock sync.Mutex

	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex
}

// send is used to send an object using the MsgPack encoding. send
// is serialized to prevent write overlaps, while properly buffering.
func (c *RPCClient) send(header *requestHeader, obj interface{}) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	if c.shutdown {
		return clientClosed
	}

	if err := c.enc.Encode(header); err != nil {
		return err
	}

	if obj != nil {
		if err := c.enc.Encode(obj); err != nil {
			return err
		}
	}

	if err := c.writer.Flush(); err != nil {
		return err
	}

	return nil
}

// NewRPCClient is used to create a new RPC client given the address.
// This will properly dial, handshake, and start listening
func NewRPCClient(addr string) (*RPCClient, error) {
	// Try to dial to serf
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// Create the client
	client := &RPCClient{
		seq:        0,
		conn:       conn.(*net.TCPConn),
		reader:     bufio.NewReader(conn),
		writer:     bufio.NewWriter(conn),
		dispatch:   make(map[uint64]seqHandler),
		shutdownCh: make(chan struct{}),
	}
	client.dec = codec.NewDecoder(client.reader,
		&codec.MsgpackHandle{RawToString: true, WriteExt: true})
	client.enc = codec.NewEncoder(client.writer,
		&codec.MsgpackHandle{RawToString: true, WriteExt: true})
	go client.listen()

	// Do the initial handshake
	if err := client.handshake(); err != nil {
		client.Close()
		return nil, err
	}
	return client, err
}

// Close is used to free any resources associated with the client
func (c *RPCClient) Close() error {
	c.shutdownLock.Lock()
	defer c.shutdownLock.Unlock()

	if !c.shutdown {
		c.shutdown = true
		close(c.shutdownCh)
		c.deregisterAll()
		return c.conn.Close()
	}
	return nil
}

// Register is used to instruct watchdog to monitor a new process. It returns
// a list of process names the were successfully added.
func (c *RPCClient) Register(configPaths []string, watchPaths, startOnLoad bool) ([]string, error) {
	header := requestHeader{
		Command: registerCommand,
		Seq:     c.getSeq(),
	}
	req := registerRequest{
		StartOnLoad: startOnLoad,
		ConfigPaths: configPaths,
		WatchPaths:  watchPaths,
	}
	var resp registerResponse

	err := c.genericRPC(&header, &req, &resp)
	return resp.Names, err
}

// handshake is used to perform the initial handshake on connect
func (c *RPCClient) handshake() error {
	header := requestHeader{
		Command: handshakeCommand,
		Seq:     c.getSeq(),
	}
	req := handshakeRequest{
		Version: MaxIPCVersion,
	}
	return c.genericRPC(&header, &req, nil)
}

// genericRPC is used to send a request and wait for an
// errorSequenceResponse, potentially returning an error
func (c *RPCClient) genericRPC(header *requestHeader, req interface{}, resp interface{}) error {
	// Setup a response handler
	errCh := make(chan error, 1)
	handler := func(respHeader *responseHeader) {
		if resp != nil {
			err := c.dec.Decode(resp)
			if err != nil {
				errCh <- err
				return
			}
		}
		errCh <- strToError(respHeader.Error)
	}
	c.handleSeq(header.Seq, &seqCallback{handler: handler})
	defer c.deregisterHandler(header.Seq)

	// Send the request
	if err := c.send(header, req); err != nil {
		return err
	}

	// Wait for a response
	select {
	case err := <-errCh:
		return err
	case <-c.shutdownCh:
		return clientClosed
	}
}

// strToError converts a string to an error if not blank
func strToError(s string) error {
	if s != "" {
		return fmt.Errorf(s)
	}
	return nil
}

// getSeq returns the next sequence number in a safe manner
func (c *RPCClient) getSeq() uint64 {
	return atomic.AddUint64(&c.seq, 1)
}

// deregisterAll is used to deregister all handlers
func (c *RPCClient) deregisterAll() {
	c.dispatchLock.Lock()
	defer c.dispatchLock.Unlock()

	for _, seqH := range c.dispatch {
		seqH.Cleanup()
	}
	c.dispatch = make(map[uint64]seqHandler)
}

// deregisterHandler is used to deregister a handler
func (c *RPCClient) deregisterHandler(seq uint64) {
	c.dispatchLock.Lock()
	seqH, ok := c.dispatch[seq]
	delete(c.dispatch, seq)
	c.dispatchLock.Unlock()

	if ok {
		seqH.Cleanup()
	}
}

// handleSeq is used to setup a handlerto wait on a response for
// a given sequence number.
func (c *RPCClient) handleSeq(seq uint64, handler seqHandler) {
	c.dispatchLock.Lock()
	defer c.dispatchLock.Unlock()
	c.dispatch[seq] = handler
}

// respondSeq is used to respond to a given sequence number
func (c *RPCClient) respondSeq(seq uint64, respHeader *responseHeader) {
	c.dispatchLock.Lock()
	seqL, ok := c.dispatch[seq]
	c.dispatchLock.Unlock()

	// Get a registered listener, ignore if none
	if ok {
		seqL.Handle(respHeader)
	}
}

// listen is used to processes data coming over the IPC channel,
// and wrote it to the correct destination based on seq no
func (c *RPCClient) listen() {
	defer c.Close()
	var respHeader responseHeader
	for {
		if err := c.dec.Decode(&respHeader); err != nil {
			if !c.shutdown {
				log.Printf("[ERR] agent.client: Failed to decode response header: %v", err)
			}
			break
		}
		c.respondSeq(respHeader.Seq, &respHeader)
	}
}
