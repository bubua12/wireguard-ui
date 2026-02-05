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

	handshakes, _ := wg.GetPeerHandshakes(server.Name)

	result := make([]PeerStatus, 0)
	for _, peer := range peers {
		online := false
		if handshakes != nil {
			if ts, ok := handshakes[peer.PublicKey]; ok {
				online = wg.IsPeerOnline(ts)
			}
		}
		result = append(result, PeerStatus{
			PublicKey: peer.PublicKey,
			Online:    online,
		})
	}

	c.JSON(http.StatusOK, result)
}
