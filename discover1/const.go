package discover

import "time"

// 网络传播类型
type NetType byte

const (
	UNICAST   NetType = iota
	MULTICAST
	BROADCAST
)

const DELIMITER = "\n\r"
const MAX_BUF_LEN = 1024 * 16

// Timeouts
const (
	respTimeout = 500 * time.Millisecond
	sendTimeout = 500 * time.Millisecond
	expiration  = 20 * time.Second

	ntpFailureThreshold = 32               // Continuous timeouts after which to check NTP
	ntpWarningCooldown  = 10 * time.Minute // Minimum amount of time to pass before repeating NTP warning
	driftThreshold      = 10 * time.Second // Allowed clock drift before warning user
)
