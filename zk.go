package main

import (
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/xgfone/go-tools/lifecycle"
)

// ZkLoggerFunc is a function wrapper of Zk Logger, which converts a function
// to the type of zk.Logger.
type ZkLoggerFunc func(string, ...interface{})

// Printf implements the interface of zk.Logger.
func (l ZkLoggerFunc) Printf(format string, args ...interface{}) {
	l(format, args...)
}

// ZkClient represents a client of a ZooKeeper connection.
type ZkClient struct {
	*zk.Conn
}

// NewZkClient returns a new ZkClient.
func NewZkClient(addrs []string, timeout int, logger zk.Logger) (zc ZkClient, err error) {
	c, ev, err := zk.Connect(addrs, time.Duration(timeout)*time.Second)
	if err != nil {
		return
	}

	if logger != nil {
		c.SetLogger(logger)
	}

	lifecycle.Register(func() { c.Close() })
	zc.Conn = c
	go func() {
		for {
			if _, ok := <-ev; !ok {
				lifecycle.Stop()
				return
			}
		}
	}()

	return
}
