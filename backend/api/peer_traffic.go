package api

import (
	"net/http"
	"strconv"
	"time"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

func GetPeerTraffic(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	peer, err := db.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}

	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}

	live := liveStates(server)[peer.PublicKey].Transfer
	report, err := db.GetPeerTrafficReport(peer.PublicKey, live, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

type trafficHistoryResponse struct {
	Days []db.TrafficDay `json:"days"`
	Rx   int64           `json:"rx"`
	Tx   int64           `json:"tx"`
}

// GetTrafficHistory returns the traffic of all peers per day, oldest first,
// with empty days filled in. ?days= picks the window (default 14).
func GetTrafficHistory(c *gin.Context) {
	days, err := strconv.Atoi(c.DefaultQuery("days", "14"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 days 参数"})
		return
	}

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

	history, err := db.GetTrafficHistory(peerKeys(peers), wg.Transfers(liveStates(server)), days, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := trafficHistoryResponse{Days: history}
	for _, d := range history {
		resp.Rx += d.Rx
		resp.Tx += d.Tx
	}
	c.JSON(http.StatusOK, resp)
}
