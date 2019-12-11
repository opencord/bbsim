/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package devices

import (
	"bytes"
	"os/exec"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/bbsim/types"
	log "github.com/sirupsen/logrus"
)

var (
	nniLogger    = log.WithFields(log.Fields{"module": "NNI"})
	dhcpServerIp = "192.168.254.1"
)

type Executor interface {
	Command(name string, arg ...string) Runnable
}

type DefaultExecutor struct{}

func (d DefaultExecutor) Command(name string, arg ...string) Runnable {
	return exec.Command(name, arg...)
}

type Runnable interface {
	Run() error
}

var executor = DefaultExecutor{}

type NniPort struct {
	// BBSIM Internals
	ID           uint32
	nniVeth      string
	upstreamVeth string

	// PON Attributes
	OperState *fsm.FSM
	Type      string
}

func CreateNNI(olt *OltDevice) (NniPort, error) {
	nniPort := NniPort{
		ID:           uint32(0),
		nniVeth:      "nni",
		upstreamVeth: "upstream",
		OperState: getOperStateFSM(func(e *fsm.Event) {
			oltLogger.Debugf("Changing NNI OperState from %s to %s", e.Src, e.Dst)
		}),
		Type: "nni",
	}
	createNNIPair(executor, olt, &nniPort)
	return nniPort, nil
}

// sendNniPacket will send a packet out of the NNI interface.
// We will send upstream only DHCP packets and drop anything else
func (n *NniPort) sendNniPacket(packet gopacket.Packet) error {
	isDhcp := packetHandlers.IsDhcpPacket(packet)
	isLldp := packetHandlers.IsLldpPacket(packet)

	if isDhcp == false && isLldp == false {
		nniLogger.WithFields(log.Fields{
			"packet": packet,
		}).Trace("Dropping NNI packet as it's not DHCP")
		return nil
	}

	if isDhcp {
		packet, err := packetHandlers.PopDoubleTag(packet)
		if err != nil {
			nniLogger.WithFields(log.Fields{
				"packet": packet,
			}).Errorf("Can't remove double tags from packet: %v", err)
			return err
		}

		handle, err := getVethHandler(n.nniVeth)
		if err != nil {
			return err
		}

		err = handle.WritePacketData(packet.Data())
		if err != nil {
			nniLogger.WithFields(log.Fields{
				"packet": packet,
			}).Errorf("Failed to send packet out of the NNI: %s", err)
			return err
		}

		nniLogger.Infof("Sent packet out of NNI")
	} else if isLldp {
		// TODO rework this when BBSim supports data-plane packets
		nniLogger.Trace("Received LLDP Packet, ignoring it")
	}
	return nil
}

//createNNIBridge will create a veth bridge to fake the connection between the NNI port
//and something upstream, in this case a DHCP server.
//It is also responsible to start the DHCP server itself
func createNNIPair(executor Executor, olt *OltDevice, nniPort *NniPort) error {

	if err := executor.Command("ip", "link", "add", nniPort.nniVeth, "type", "veth", "peer", "name", nniPort.upstreamVeth).Run(); err != nil {
		nniLogger.Errorf("Couldn't create veth pair between %s and %s", nniPort.nniVeth, nniPort.upstreamVeth)
		return err
	}

	if err := setVethUp(executor, nniPort.nniVeth); err != nil {
		return err
	}

	if err := setVethUp(executor, nniPort.upstreamVeth); err != nil {
		return err
	}

	// TODO should be moved out of this function in case there are multiple NNI interfaces.
	// Only one DHCP server should be running and listening on all NNI interfaces
	if err := startDHCPServer(nniPort.upstreamVeth, dhcpServerIp); err != nil {
		return err
	}

	return nil
}

func deleteNNIPair(executor Executor, nniPort *NniPort) error {
	if err := executor.Command("ip", "link", "del", nniPort.nniVeth).Run(); err != nil {
		nniLogger.Errorf("Couldn't delete veth pair between %s and %s", nniPort.nniVeth, nniPort.upstreamVeth)
		return err
	}
	return nil
}

// NewVethChan returns a new channel for receiving packets over the NNI interface
func (n *NniPort) NewVethChan() (chan *types.PacketMsg, *pcap.Handle, error) {
	ch, handle, err := listenOnVeth(n.nniVeth)
	if err != nil {
		return nil, nil, err
	}
	return ch, handle, err
}

// setVethUp is responsible to activate a virtual interface
func setVethUp(executor Executor, vethName string) error {
	if err := executor.Command("ip", "link", "set", vethName, "up").Run(); err != nil {
		nniLogger.Errorf("Couldn't change interface %s state to up: %v", vethName, err)
		return err
	}
	return nil
}

var startDHCPServer = func(upstreamVeth string, dhcpServerIp string) error {
	// TODO the DHCP server should support multiple interfaces
	if err := exec.Command("ip", "addr", "add", dhcpServerIp, "dev", upstreamVeth).Run(); err != nil {
		nniLogger.Errorf("Couldn't assing ip %s to interface %s: %v", dhcpServerIp, upstreamVeth, err)
		return err
	}

	if err := setVethUp(executor, upstreamVeth); err != nil {
		return err
	}

	dhcp := "/usr/local/bin/dhcpd"
	conf := "/etc/dhcp/dhcpd.conf" // copied in the container from configs/dhcpd.conf
	logfile := "/tmp/dhcplog"
	var stderr bytes.Buffer
	cmd := exec.Command(dhcp, "-cf", conf, upstreamVeth, "-tf", logfile, "-4")
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		nniLogger.Errorf("Fail to start DHCP Server: %s, %s", err, stderr.String())
		return err
	}
	nniLogger.Info("Successfully activated DHCP Server")
	return nil
}

func getVethHandler(vethName string) (*pcap.Handle, error) {
	var (
		device            = vethName
		snapshotLen int32 = 1518
		promiscuous       = false
		timeout           = pcap.BlockForever
	)
	handle, err := pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if err != nil {
		nniLogger.Errorf("Can't retrieve handler for interface %s", vethName)
		return nil, err
	}
	return handle, nil
}

var listenOnVeth = func(vethName string) (chan *types.PacketMsg, *pcap.Handle, error) {

	handle, err := getVethHandler(vethName)
	if err != nil {
		return nil, nil, err
	}

	channel := make(chan *types.PacketMsg, 1024)

	go func() {
		nniLogger.Info("Start listening on NNI for packets")
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {

			if !packetHandlers.IsIncomingPacket(packet) {
				nniLogger.Tracef("Ignoring packet as it's going out")
				continue
			}

			nniLogger.WithFields(log.Fields{
				"packet": packet.Dump(),
			}).Tracef("Received packet on NNI Port")
			pkt := types.PacketMsg{
				Pkt: packet,
			}
			channel <- &pkt
		}
		nniLogger.Info("Stop listening on NNI for packets")
	}()

	return channel, handle, nil
}
