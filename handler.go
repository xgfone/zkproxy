package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/xgfone/go-tools/nets/https"
)

// Handler is a HTTP handler.
type Handler struct {
	zk     ZkClient
	prefix string
	acl    []zk.ACL
}

// NewHandler returns a new Handler.
func NewHandler(prefix string, zkClient ZkClient) Handler {
	return Handler{
		prefix: strings.TrimRight(prefix, "/"),
		zk:     zkClient,
		acl:    []zk.ACL{zk.ACL{Perms: 0x1f, Scheme: "world", ID: "anyone"}},
	}
}

// Path fixes the path and returns the final path.
func (h Handler) Path(path string) string {
	if h.prefix == "" {
		return path
	}
	if path[0] == '/' {
		return h.prefix + path
	}
	return fmt.Sprintf("%s/%s", h.prefix, path)
}

func (h Handler) toJSON(m map[string]interface{}, s *zk.Stat) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{}, 11)
	}
	m["czxid"] = s.Czxid
	m["mzxid"] = s.Mzxid
	m["ctime"] = s.Ctime
	m["mtime"] = s.Mtime
	m["version"] = s.Version
	m["cversion"] = s.Cversion
	m["aversion"] = s.Aversion
	m["ephemeral_owner"] = s.EphemeralOwner
	m["data_length"] = s.DataLength
	m["num_children"] = s.NumChildren
	m["pzxid"] = s.Pzxid
	return m
}

// HandleZk handle the request of ZooKeeper.
func (h Handler) HandleZk(w http.ResponseWriter, r *http.Request) (code int, resp []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			code = http.StatusBadRequest
			err, _ = e.(error)
		}
	}()

	info := make(map[string]interface{})
	if err = https.DecodeJSON(r, 1024*1024, &info); err != nil {
		code = http.StatusBadRequest
		return
	}

	code = http.StatusOK
	cmd := info["cmd"].(string)
	cmd = strings.Replace(strings.ToLower(cmd), "-", "_", -1)
	switch cmd {
	case "add_auth_info":
		err = h.AddAuthInfo(info)
	case "create":
		return h.Create(info)
	case "delete":
		return h.Delete(info)
	case "exists":
		resp, err = h.Exists(info)
	case "get_children":
		return h.GetChildren(info)
	case "get_data":
		return h.GetData(info)
	case "set_data":
		return h.SetData(info)
	case "get_acl":
		return h.GetACL(info)
	case "set_acl":
		return h.SetACL(info)
	// case "exists_watch", "get_children_watch", "get_data_watch", "multi":
	default:
		code = http.StatusNotImplemented
		err = fmt.Errorf("The cmd %s is not implemented", cmd)
	}

	return
}

// AddAuthInfo adds the auth into the ZK cluster.
func (h Handler) AddAuthInfo(info map[string]interface{}) (err error) {
	return h.zk.AddAuth(info["scheme"].(string), []byte(info["auth"].(string)))
}

// Create creates the path as a new.
func (h Handler) Create(info map[string]interface{}) (code int, resp []byte, err error) {
	path := h.Path(info["path"].(string))
	data := []byte(info["data"].(string))

	var acls []zk.ACL
	if info["acl"] == nil {
		acls = h.acl
	} else {
		_acls := info["acl"].([]interface{})
		acls = make([]zk.ACL, len(_acls))
		for i, acl := range _acls {
			_acl := acl.(map[string]interface{})
			acls[i].Perms = int32(_acl["perms"].(float64))
			acls[i].Scheme = _acl["scheme"].(string)
			acls[i].ID = _acl["id"].(string)
		}
	}

	var ephemeral bool
	if info["ephemeral"] != nil {
		ephemeral = info["ephemeral"].(bool)
	}

	var sequential bool
	if info["sequential"] != nil {
		sequential = info["sequential"].(bool)
	}

	if ephemeral && sequential {
		path, err = h.zk.CreateProtectedEphemeralSequential(path, data, acls)
	} else {
		var flags int32
		if ephemeral {
			flags |= zk.FlagEphemeral
		}
		if sequential {
			flags |= zk.FlagSequence
		}
		path, err = h.zk.Create(path, data, flags, acls)
	}

	if err == nil {
		resp, err = json.Marshal(map[string]interface{}{"path": path})
	} else if err == zk.ErrNodeExists {
		code = http.StatusNotAcceptable
	} else if err == zk.ErrNoNode {
		code = http.StatusNotFound
	}

	return
}

// Delete deletes the path.
func (h Handler) Delete(info map[string]interface{}) (code int, resp []byte, err error) {
	path := h.Path(info["path"].(string))
	version := int32(info["version"].(float64))
	code = http.StatusOK
	if err = h.zk.Delete(path, version); err == zk.ErrNoNode {
		code = http.StatusNotFound
	} else if err == zk.ErrBadVersion {
		code = http.StatusNotAcceptable
	}
	return
}

// Exists returns true if the path exists, or false if not.
func (h Handler) Exists(info map[string]interface{}) (resp []byte, err error) {
	yes, s, err := h.zk.Exists(h.Path(info["path"].(string)))
	if err == nil {
		if yes {
			resp, err = json.Marshal(h.toJSON(map[string]interface{}{"exist": true}, s))
		} else {
			resp, err = json.Marshal(map[string]interface{}{"exist": false})
		}
	}
	return
}

// GetChildren returns the children information of the path.
func (h Handler) GetChildren(info map[string]interface{}) (code int, resp []byte, err error) {
	children, s, err := h.zk.Children(h.Path(info["path"].(string)))
	if err == nil && s != nil {
		resp, err = json.Marshal(h.toJSON(map[string]interface{}{"children": children}, s))
	} else if err == zk.ErrNoNode {
		code = http.StatusNotFound
	}
	return
}

// GetData returns the data information of the path.
func (h Handler) GetData(info map[string]interface{}) (code int, resp []byte, err error) {
	data, s, err := h.zk.Get(h.Path(info["path"].(string)))
	if err == nil && s != nil {
		resp, err = json.Marshal(h.toJSON(map[string]interface{}{"data": string(data)}, s))
	} else if err == zk.ErrNoNode {
		code = http.StatusNotFound
	}
	return
}

// SetData sets the data information of the path.
func (h Handler) SetData(info map[string]interface{}) (code int, resp []byte, err error) {
	path := h.Path(info["path"].(string))
	data := info["data"].(string)
	version := int32(info["version"].(float64))
	s, err := h.zk.Set(path, []byte(data), version)
	if err == nil && s != nil {
		resp, err = json.Marshal(h.toJSON(nil, s))
	} else if err == zk.ErrNoNode {
		code = http.StatusNotFound
	} else if err == zk.ErrBadVersion {
		code = http.StatusNotAcceptable
	}
	return
}

// GetACL returns the ACL of the path.
func (h Handler) GetACL(info map[string]interface{}) (code int, resp []byte, err error) {
	acls, s, err := h.zk.GetACL(h.Path(info["path"].(string)))
	if err == nil && s != nil {
		_acls := make([]map[string]interface{}, len(acls))
		for i, acl := range acls {
			_acls[i] = map[string]interface{}{
				"id":     acl.ID,
				"perms":  acl.Perms,
				"scheme": acl.Scheme,
			}
		}
		resp, err = json.Marshal(h.toJSON(map[string]interface{}{"acl": _acls}, s))
	} else if err == zk.ErrNoNode {
		code = http.StatusNotFound
	}
	return
}

// SetACL sets the ACL of the path.
func (h Handler) SetACL(info map[string]interface{}) (code int, resp []byte, err error) {
	path := h.Path(info["path"].(string))
	version := int32(info["version"].(float64))
	_acls := info["acl"].([]interface{})

	acls := make([]zk.ACL, len(_acls))
	for i, acl := range _acls {
		_acl := acl.(map[string]interface{})
		acls[i].Perms = int32(_acl["perms"].(float64))
		acls[i].Scheme = _acl["scheme"].(string)
		acls[i].ID = _acl["id"].(string)
	}

	s, err := h.zk.SetACL(path, acls, version)
	if err == nil && s != nil {
		resp, err = json.Marshal(h.toJSON(nil, s))
	} else if err == zk.ErrNoNode {
		code = http.StatusNotFound
	} else if err == zk.ErrBadVersion {
		code = http.StatusNotAcceptable
	}

	return
}
