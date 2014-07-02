package main

import (
	"errors"
	"fmt"
	"github.com/izqui/helpers"
	"sort"
)

type PeerSlice []*Peer

func (slice PeerSlice) Len() int {

	return len(slice)
}

func (slice PeerSlice) Swap(i, j int) {

	slice[i], slice[j] = slice[j], slice[i]
}

func (slice PeerSlice) Less(i, j int) bool {

	return slice[i].Id < slice[j].Id
}

type Peer struct {
	Id             string    `json:"id"`
	Address        string    `json:"address"`
	ConnectedPeers PeerSlice `json:"connected_peers,omitempty"`
}

type Message struct {
	Id                 string `json:"id"`
	Body               string `json:"body,omitempty"`
	Origin             *Peer  `json:"origin_peer"`
	Destination        string `json:"-"`
	FinalDestinationId string `json:"destination_id"`
}

func (m *Message) AssignRandomID() {

	m.Id = helpers.SHA1([]byte(helpers.RandomString(10)))
}

func (p *Peer) String() string {

	con := ""
	if len(p.ConnectedPeers) > 0 {

		for _, c := range p.ConnectedPeers {

			if c != nil {
				con += c.String()
			}
		}
	}

	return fmt.Sprintf("%s(%s) -> [%s], ", p.Id, p.Address, con)
}

func (p *Peer) AddConnectedPeer(newPeer *Peer) error {

	if p.Id == newPeer.Id {

		return errors.New("You are trying to add yourself as a peer")
	}

	location := -1

	for i, con := range p.ConnectedPeers {

		if con.Id == newPeer.Id {

			if con.Hash() == newPeer.Hash() {

				return errors.New("Peer was already connected")
			} else {

				// Insert at this location, same id, different shit inside
				location = i
			}
		}
	}

	if location == -1 {

		p.ConnectedPeers = append(p.ConnectedPeers, newPeer)
	} else {
		p.ConnectedPeers[location] = newPeer
	}

	//Remove myself if I'm referenced in other peers
	p.removeIfPresent(p.Id)

	return nil
}

func (p *Peer) Hash() string {

	data := p.Id

	if len(p.ConnectedPeers) > 0 {

		sort.Sort(p.ConnectedPeers)

		for _, c := range p.ConnectedPeers {

			if c != nil {

				data += c.Hash()
			}
		}
	}

	return helpers.SHA1([]byte(data))
}

func (p *Peer) removeIfPresent(id string) {

	connected := p.ConnectedPeers

	for i, c := range p.ConnectedPeers {

		if c != nil {

			if c.Id == id {

				connected = remove(connected, i)
			} else {

				c.removeIfPresent(id)
			}
		}
	}

	p.ConnectedPeers = connected
}

func (p *Peer) FindNearestPeerToId(id string) *Peer {

	for _, c := range p.ConnectedPeers {

		if c.Id == id {

			return c
		}
	}

	distance := 1000
	var peer *Peer = nil

	for _, c := range p.ConnectedPeers {

		n := c.distanceToId(id)

		if n > -1 && n < distance {

			peer = c
			distance = n
		}
	}

	return peer
}

func (p Peer) distanceToId(id string) int {

	for _, c := range p.ConnectedPeers {

		if c.Id == id {

			return 1
		}

		n := c.distanceToId(id)
		if n > -1 {

			return 1 + n
		}
	}

	return -1
}

func remove(slice PeerSlice, i int) PeerSlice {

	copy(slice[i:], slice[i+1:])
	slice[len(slice)-1] = nil
	return slice[:len(slice)-1]
}
