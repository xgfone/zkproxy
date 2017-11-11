package main

import (
	"net/http"
	"syscall"

	"github.com/golang/glog"
	"github.com/xgfone/go-config"
	"github.com/xgfone/go-tools/nets/https"
	"github.com/xgfone/go-tools/signal2"
)

var (
	conf = config.Conf
)

var globalOpts = []config.Opt{
	config.Str("addr", ":8000", "The address that the HTTP server listens to."),
}

var zkOpts = []config.Opt{
	config.Strings("addrs", []string{"127.0.0.1:2181"},
		"The address list of the ZooKeeper cluster, which are separated by the comma."),
	config.Str("prefix", "", "The prefix of the path."),
	config.Int("timeout", 3, "The session timeout."),
}

func init() {
	glog.MaxSize = 1024 * 1024 * 1024
	https.ErrorLogFunc = glog.Errorf

	conf.RegisterCliOpts("", globalOpts)
	conf.RegisterCliOpts("zk", zkOpts)
}

func main() {
	if err := conf.Parse(nil); err != nil {
		glog.Exit(err)
	}

	// Handle the signals.
	go signal2.HandleSignal(syscall.SIGTERM, syscall.SIGQUIT)

	// Connect to the ZooKeeper cluster.
	zkOpts := conf.Group("zk")
	zk, err := NewZkClient(zkOpts.Strings("addrs"), zkOpts.Int("timeout"), ZkLoggerFunc(glog.Infof))
	if err != nil {
		glog.Exit(err)
	}

	// Create a HTTP handler and set the router.
	handler := NewHandler(zkOpts.String("prefix"), zk)
	http.HandleFunc("/zk", https.HandlerWrapper(handler.HandleZk))

	// Start the HTTP server.
	glog.Exit(http.ListenAndServe(conf.String("addr"), nil))
}
