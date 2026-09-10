package cluster

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type NodeRole string

const (
	RolePrimaryRender NodeRole = "RENDER_STARTER"
	RoleSecondaryVPS  NodeRole = "BUDGET_VPS"
	RoleExpansionNode NodeRole = "FUTURE_CLUSTER_NODE"
)

type ServerNode struct {
	NodeID          string
	URL             *url.URL
	Role            NodeRole
	MaxCapacityRPS  int32
	CurrentActiveRPS int32
	IsHealthy       bool
	ProxyHandler    *httputil.ReverseProxy
}

type DynamicClusterMesh struct {
	nodes []*ServerNode
	mu    sync.RWMutex
}

func NewDynamicClusterMesh() *DynamicClusterMesh {
	return &DynamicClusterMesh{
		nodes: make([]*ServerNode, 0),
	}
}

// RegisterServerNode: नया सर्वर जुड़ते ही 1 सेकंड में क्लस्टर में शामिल करना
func (cm *DynamicClusterMesh) RegisterServerNode(nodeID, rawURL string, role NodeRole, maxRPS int32) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("अमान्य सर्वर URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	node := &ServerNode{
		NodeID:          nodeID,
		URL:             parsedURL,
		Role:            role,
		MaxCapacityRPS:  maxRPS,
		CurrentActiveRPS: 0,
		IsHealthy:       true,
		ProxyHandler:    proxy,
	}

	cm.nodes = append(cm.nodes, node)
	log.Printf("🌐 [CLUSTER] नया नोड पंजीकृत: ID=%s | URL=%s | क्षमता=%d RPS", nodeID, rawURL, maxRPS)
	return nil
}

// RouteSmartTraffic: लोड के हिसाब से ट्रैफिक सबसे कम लोड वाले सर्वर पर ऑटो-स्कैल्प करना
func (cm *DynamicClusterMesh) RouteSmartTraffic(w http.ResponseWriter, r *http.Request) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var selectedNode *ServerNode
	var lowestLoadRatio float64 = 1.0

	// सबसे स्वस्थ और कम लोड वाले सर्वर का चुनाव (Least Connection / Elastic Scalping)
	for _, node := range cm.nodes {
		if !node.IsHealthy {
			continue
		}
		ratio := float64(atomic.LoadInt32(&node.CurrentActiveRPS)) / float64(node.MaxCapacityRPS)
		if ratio < lowestLoadRatio {
			lowestLoadRatio = ratio
			selectedNode = node
		}
	}

	if selectedNode == nil {
		http.Error(w, "सभी सर्वर वर्तमान में व्यस्त हैं। कृपया 5 सेकंड बाद पुनः प्रयास करें।", http.StatusServiceUnavailable)
		return
	}

	// लोड काउंटर बढ़ाना
	atomic.AddInt32(&selectedNode.CurrentActiveRPS, 1)
	defer atomic.AddInt32(&selectedNode.CurrentActiveRPS, -1)

	// ऑटो-फॉरवर्ड
	selectedNode.ProxyHandler.ServeHTTP(w, r)
}

// StartAutoScalingWatchdog: हर 5 सेकंड में सभी सर्वरों की सेहत और बैकअप की जांच
func (cm *DynamicClusterMesh) StartAutoScalingWatchdog() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		client := http.Client{Timeout: 2 * time.Second}

		for range ticker.C {
			cm.mu.RLock()
			for _, node := range cm.nodes {
				healthURL := fmt.Sprintf("%s/api/v1/health", node.URL.String())
				resp, err := client.Get(healthURL)
				if err != nil || resp.StatusCode != http.StatusOK {
					if node.IsHealthy {
						node.IsHealthy = false
						log.Printf("⚠️ [FAILOVER] सर्वर %s अनअवेलेबल! ट्रैफिक अन्य नोड्स पर ऑटो-डाइवर्ट हुआ।", node.NodeID)
					}
				} else {
					if !node.IsHealthy {
						node.IsHealthy = true
						log.Printf("✅ [RESTORED] सर्वर %s ठीक हुआ। क्लस्टर में दोबारा सक्रिय।", node.NodeID)
					}
					resp.Body.Close()
				}
			}
			cm.mu.RUnlock()
		}
	}()
}
