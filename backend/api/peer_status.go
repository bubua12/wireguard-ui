package api

import (
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

type PeerStatus struct {
	PublicKey string `json:"public_key"`
	Online    bool   `json:"online"`
	Rx        int64  `json:"rx"`
	Tx        int64  `json:"tx"`
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

	handshakes, _ := wg.GetPeerHandshakes(server.Interface)
	live, _ := wg.GetPeerTransferMap(server.Interface)
	if live == nil {
		live = map[string]wg.Transfer{}
	}

	keys := make([]string, 0, len(peers))
	for _, peer := range peers {
		keys = append(keys, peer.PublicKey)
	}
	usage, err := db.UsageForKeys(keys, live)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]PeerStatus, 0, len(peers))
	for _, peer := range peers {
		online := false
		if handshakes != nil {
			if ts, ok := handshakes[peer.PublicKey]; ok {
				online = wg.IsPeerOnline(ts)
			}
		}
		u := usage[peer.PublicKey]
		result = append(result, PeerStatus{
			PublicKey: peer.PublicKey,
			Online:    online,
			Rx:        u.Rx,
			Tx:        u.Tx,
		})
	}

	c.JSON(http.StatusOK, result)
}
