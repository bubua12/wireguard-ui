package api

import (
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

// PeerStatus is the live state of one peer. Rx/Tx are accumulated totals from
// the server's side: rx is what the peer uploaded, tx is what it downloaded.
type PeerStatus struct {
	PublicKey       string `json:"public_key"`
	Online          bool   `json:"online"`
	Rx              int64  `json:"rx"`
	Tx              int64  `json:"tx"`
	Endpoint        string `json:"endpoint,omitempty"`
	LatestHandshake int64  `json:"latest_handshake"`
	// HandshakeAge is the seconds since the last handshake, -1 for never, and
	// null when the interface is down so nothing is known. It is computed here
	// so the UI does not depend on the browser's clock.
	HandshakeAge *int64 `json:"handshake_age"`
}

func GetPeersStatus(c *gin.Context) {
	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}

	peers, err := db.GetPeersByServer(server.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	states := liveStates(server)
	usage, err := db.UsageForKeys(peerKeys(peers), wg.Transfers(states))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]PeerStatus, 0, len(peers))
	for _, peer := range peers {
		st := states[peer.PublicKey]
		u := usage[peer.PublicKey]
		var age *int64
		if states != nil {
			a := wg.HandshakeAge(st.LatestHandshake)
			age = &a
		}
		result = append(result, PeerStatus{
			PublicKey:       peer.PublicKey,
			Online:          wg.IsPeerOnline(st.LatestHandshake),
			Rx:              u.Rx,
			Tx:              u.Tx,
			Endpoint:        st.Endpoint,
			LatestHandshake: st.LatestHandshake,
			HandshakeAge:    age,
		})
	}

	c.JSON(http.StatusOK, result)
}
