package ethernet

import "sync"

// DemuxOutput is a function that accepts incoming Ethernet packets.
type DemuxOutput func(Packet)

// Demux will demultiplex incoming Ethernet packets.
type Demux struct {
	sync.RWMutex
	outputs map[EtherType]DemuxOutput
}

// NewDemux creates an Ethernet demultiplexer with a default output function.
func NewDemux(nic NIC, defaultOutput DemuxOutput) *Demux {
	demux := &Demux{
		outputs: make(map[EtherType]DemuxOutput),
	}

	demux.outputs[EtherType(0)] = defaultOutput
	go demux.receiveAll(nic)

	return demux
}

// SetOutput sets an output function for a specific EtherType.
func (demux *Demux) SetOutput(etherType EtherType, output DemuxOutput) {
	if etherType.IsLength() {
		panic("must be a true EtherType, not a payload length")
	}

	demux.RWMutex.Lock()
	demux.outputs[etherType] = output
	demux.RWMutex.Unlock()
}

func (demux *Demux) receiveAll(nic NIC) {
	for p := range nic.Receive() {

		demux.RWMutex.RLock()
		output, ok := demux.outputs[p.EtherType]
		if !ok || p.EtherType.IsLength() {
			output = demux.outputs[EtherType(0)]
		}
		demux.RWMutex.RUnlock()

		output(p)
	}
}