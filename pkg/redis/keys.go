package redis

import "time"

const (
	KeyAuthToken      = "auth:token:%d"      // auth:token:{userId} TTL=2h
	KeyAuthRefresh    = "auth:refresh:%d:%s"  // auth:refresh:{userId}:{tokenId} TTL=7d
	KeyPermission     = "permission:%d"       // permission:{userId} TTL=30min
	KeyConfig         = "config:%s"           // config:{key} TTL=10min
	KeyConfigGroup    = "config:%s:*"         // config:{group}:* pattern for scan
	KeyHeartbeat      = "heartbeat:device:%d" // heartbeat:device:{id} TTL=30s
	KeySessionEntity  = "session:entity:%s"   // session:entity:{sid} TTL=5s
	KeyLoginFail      = "login_fail:%d"       // login_fail:{userId} TTL=300s
	KeySessionLock    = "session_lock:%d"     // session_lock:{sessionId} TTL=10s
	KeySimDedup       = "sim:dedup:%s:%d"    // sim:dedup:{entityId}:{seq} TTL=60s
	KeySituationPos   = "situation:pos:%d"    // situation:pos:{entityId} TTL=5s
	KeyConfigChanged  = "config:changed"      // Pub/Sub channel for config changes
	KeySessionStatus  = "session_status"      // Pub/Sub channel for session status
)

var TTL = struct {
	AuthToken   time.Duration
	AuthRefresh time.Duration
	Permission  time.Duration
	Config      time.Duration
	Heartbeat   time.Duration
	SessionEnt  time.Duration
	LoginFail   time.Duration
	SessionLock time.Duration
	SimDedup    time.Duration
	SitPosition time.Duration
}{
	AuthToken:   2 * time.Hour,
	AuthRefresh: 7 * 24 * time.Hour,
	Permission:  30 * time.Minute,
	Config:      10 * time.Minute,
	Heartbeat:   30 * time.Second,
	SessionEnt:  5 * time.Second,
	LoginFail:   5 * time.Minute,
	SessionLock: 10 * time.Second,
	SimDedup:    60 * time.Second,
	SitPosition: 5 * time.Second,
}
