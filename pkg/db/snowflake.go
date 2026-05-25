package db

import (
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const (
	epoch        = 1700000000000 // 2023-11-01 00:00:00 UTC in ms
	workerBits   = 10
	sequenceBits = 12
	maxWorkerID  = -1 ^ (-1 << workerBits)
	maxSequence  = -1 ^ (-1 << sequenceBits)

	workerShift   = sequenceBits
	timestampShift = sequenceBits + workerBits
)

type Snowflake struct {
	mu        sync.Mutex
	timestamp int64
	workerID  int64
	sequence  int64
}

var (
	instance *Snowflake
	once     sync.Once
)

func GetSnowflake() *Snowflake {
	once.Do(func() {
		workerID := autoWorkerID()
		if workerID < 0 || workerID > maxWorkerID {
			panic(fmt.Sprintf("snowflake worker ID %d out of range [0,%d]", workerID, maxWorkerID))
		}
		instance = &Snowflake{
			timestamp: 0,
			workerID:  workerID,
			sequence:  0,
		}
	})
	return instance
}

func autoWorkerID() int64 {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	hash := md5.Sum([]byte(hostname))
	id := int64(hash[0])<<8 | int64(hash[1])

	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					id = (id + int64(ip4[2])<<8 + int64(ip4[3])) & maxWorkerID
					break
				}
			}
		}
	}
	return id
}

func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli() - epoch
	if now < s.timestamp {
		panic("clock moved backwards")
	}

	if now == s.timestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			for now <= s.timestamp {
				now = time.Now().UnixMilli() - epoch
			}
		}
	} else {
		s.sequence = 0
	}

	s.timestamp = now
	return (now << timestampShift) | (s.workerID << workerShift) | s.sequence
}
