package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
	"wireguard-ui/db"
	"wireguard-ui/model"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

// mutateMu serialises every handler that changes the database and then
// rewrites or applies the wg config, so two requests cannot interleave the
// persist and apply steps and leave a stale config on disk.
var mutateMu sync.Mutex

const maxPeerNameLen = 64

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return 0, false
	}
	return id, true
}

func persistPeers(server *model.Server) error {
	peers, err := db.GetPeersByServer(server.ID)
	if err != nil {
		return err
	}
	return wg.PersistServerConfig(server, peers)
}

// applyChange rewrites the config file and then runs apply against the live
// interface (nil to skip). The database change already succeeded by the time
// this runs, so the first failure comes back as a warning for the caller to
// surface rather than as a failed request.
func applyChange(server *model.Server, apply func() error) error {
	if err := persistPeers(server); err != nil {
		log.Printf("persist wg config: %v", err)
		return err
	}
	if apply == nil {
		return nil
	}
	if err := apply(); err != nil {
		log.Printf("apply to live interface: %v", err)
		return err
	}
	return nil
}

func applyWarning(c *gin.Context, payload gin.H, err error) {
	if err != nil {
		payload["warning"] = err.Error()
	}
	c.JSON(http.StatusOK, payload)
}

// peerResponse is a peer plus an optional apply warning, flattened into one
// JSON object.
type peerResponse struct {
	*model.Peer
	Warning string `json:"warning,omitempty"`
}

func newPeerResponse(peer *model.Peer, warning error) peerResponse {
	r := peerResponse{Peer: peer}
	if warning != nil {
		r.Warning = warning.Error()
	}
	return r
}

func peerKeys(peers []model.Peer) []string {
	keys := make([]string, 0, len(peers))
	for _, p := range peers {
		keys = append(keys, p.PublicKey)
	}
	return keys
}

// liveStates returns the kernel's view of the interface, or nil when it is down
// or unreadable; callers then fall back to the stored totals.
func liveStates(server *model.Server) map[string]wg.PeerState {
	if !wg.InterfaceExists(server.Interface) {
		return nil
	}
	states, err := wg.GetPeerStates(server.Interface)
	if err != nil {
		return nil
	}
	return states
}

func validatePeerName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("名称不能为空")
	}
	if utf8.RuneCountInString(name) > maxPeerNameLen {
		return "", fmt.Errorf("名称不能超过 %d 个字符", maxPeerNameLen)
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("名称不能包含控制字符")
		}
	}
	return name, nil
}

func validateKeepalive(n int) error {
	if n < 0 || n > 65535 {
		return fmt.Errorf("保活间隔应在 0-65535 秒之间（0 表示关闭）")
	}
	return nil
}

func defaultsForServer(s *model.Server) {
	if s.Interface == "" {
		s.Interface = "wg0"
	}
	if s.ListenPort == 0 {
		s.ListenPort = 51820
	}
	if s.DNS == "" {
		s.DNS = "8.8.8.8"
	}
	if s.MTU == 0 {
		s.MTU = 1420
	}
	if s.Name == "" {
		s.Name = s.Interface
	}
}

// validateServerFields checks the user-supplied server settings and returns the
// DNS list in its canonical "a, b" form.
func validateServerFields(iface, address, endpoint, dns string, port, mtu int) (string, error) {
	if err := wg.ValidateInterfaceName(iface); err != nil {
		return "", err
	}
	if err := wg.ValidateServerAddress(address); err != nil {
		return "", err
	}
	if err := wg.ValidateEndpoint(endpoint); err != nil {
		return "", err
	}
	if err := wg.ValidateListenPort(port); err != nil {
		return "", err
	}
	if err := wg.ValidateMTU(mtu); err != nil {
		return "", err
	}
	if err := wg.ValidateDNS(dns); err != nil {
		return "", err
	}
	return wg.NormalizeDNS(dns), nil
}
