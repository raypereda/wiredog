package httpassembler

// The code in this file is based on the example from the gopacket package
// https://github.com/google/gopacket/blob/master/examples/httpassembly/main.go

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/pkg/errors"
)

type httpStreamFactory struct {
	requests chan<- *http.Request
}

func (h *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	hstream := &httpStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
		requests:  h.requests,
	}
	go hstream.run()
	return &hstream.r
}

type httpStream struct {
	requests       chan<- *http.Request
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
}

func (h *httpStream) run() {
	buf := bufio.NewReader(&h.r)
	for {
		req, err := http.ReadRequest(buf)
		if err == io.EOF {
			return
		} else if err != nil {
			log.Println("error reading stream", h.net, h.transport, ":", err)
		} else {
			req.Body.Close()
			h.requests <- req
		}
	}
}

// HTTPAssembler assembles HTTP requests from a local network packets.
type HTTPAssembler struct {
	device    string
	handle    *pcap.Handle
	assembler *tcpassembly.Assembler
}

// NewHTTPAssembler creates a new source of assembled HTTP requests from packets.
// The requests are passed back via the provided channel.
func NewHTTPAssembler(device string, requests chan<- *http.Request) (*HTTPAssembler, error) {
	maxReadSize := int32(0) // unbounded
	promiscuous := true     // put device in mode to read all packets on all devices
	timeout := pcap.BlockForever
	handle, err := pcap.OpenLive(device, maxReadSize, promiscuous, timeout)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening pcap handle")
	}

	filter := "tcp and dst port 80" // BPF filter for pcap looking for HTTP requests
	err = handle.SetBPFFilter(filter)
	if err != nil {
		return nil, errors.Wrapf(err, "error setting BPF filter")
	}

	streamFactory := &httpStreamFactory{requests: requests}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	return &HTTPAssembler{
		device:    device,
		handle:    handle,
		assembler: assembler,
	}, nil
}

// Run starts the HTTP request sniffing process.
func (r *HTTPAssembler) Run() {
	defer r.handle.Close()
	packetSource := gopacket.NewPacketSource(r.handle, r.handle.LinkType())
	packets := packetSource.Packets()

	ticker := time.Tick(time.Minute)
	for {
		select {
		case packet := <-packets:
			if packet == nil {
				return
			}
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil ||
				packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				continue
			}

			tcp := packet.TransportLayer().(*layers.TCP)
			r.assembler.AssembleWithTimestamp(
				packet.NetworkLayer().NetworkFlow(),
				tcp, time.Now())
			// tcp, packet.Metadata().Timestamp)
			// The metadata timestamp is often empty.

			// Experiment: Identify HTTP requests by looking in payload.
			// Decided against it and rely on Berkeley Packet Filter instead.
			// Search for a string inside the payload. 
			// applicationLayer := packet.ApplicationLayer()
			// if applicationLayer != nil && strings.Contains(string(applicationLayer.Payload()), "HTTP") {
			// 	fmt.Println("HTTP found!")
			// }

		case <-ticker:
			// Every minute, flush connections that haven't seen
			// activity in the past 2 minutes.
			r.assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		}
	}
}
