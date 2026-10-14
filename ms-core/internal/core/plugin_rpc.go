package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/rpc"

	hcPlugin "github.com/hashicorp/go-plugin"
)

// HandshakeConfig is the shared configuration that both host and subprocess
// must present to authenticate the connection.
// Both sides must use identical values — mismatches are rejected immediately.
var HandshakeConfig = hcPlugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "MS_CORE_PLUGIN",
	MagicCookieValue: "ms-core-plugin-v1",
}

// PluginMap is used by the host when dispensing a plugin client.
// The Impl field is intentionally nil here; it is populated on the
// subprocess side inside ServePlugin().
var PluginMap = map[string]hcPlugin.Plugin{
	"plugin": &PluginNetRPC{},
}

// ServePlugin starts the net/rpc server inside a plugin subprocess.
// Call this at the end of each plugin's main() after any initialization
// (e.g. database connection).  This call blocks until the host kills the process.
func ServePlugin(impl Plugin) {
	hcPlugin.Serve(&hcPlugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]hcPlugin.Plugin{
			"plugin": &PluginNetRPC{Impl: impl},
		},
	})
}

// ---------------------------------------------------------------------------
// PluginNetRPC — hashicorp/go-plugin bridge over net/rpc
// ---------------------------------------------------------------------------

// PluginNetRPC bridges core.Plugin over net/rpc.
// Server() runs inside the subprocess; Client() runs in the host.
type PluginNetRPC struct {
	// Impl is only set on the subprocess (server) side.
	Impl Plugin
}

// Server returns the net/rpc handler that runs inside the plugin subprocess.
func (p *PluginNetRPC) Server(*hcPlugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

// Client returns the proxy that runs in the host and implements core.Plugin.
// Name and Version are fetched once and cached to avoid per-call RPC overhead.
func (p *PluginNetRPC) Client(_ *hcPlugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	cl := &RPCClient{client: c}
	var nr StringReply
	if err := c.Call("Plugin.Name", struct{}{}, &nr); err == nil {
		cl.cachedName = nr.Value
	}
	var vr StringReply
	if err := c.Call("Plugin.Version", struct{}{}, &vr); err == nil {
		cl.cachedVersion = vr.Value
	}
	return cl, nil
}

// ---------------------------------------------------------------------------
// Wire types
// ---------------------------------------------------------------------------

// StringReply carries a single string value over the wire.
type StringReply struct{ Value string }

// ErrorReply carries an optional error string over the wire.
// An empty Error field means success.
type ErrorReply struct{ Error string }

// ExecuteArgs carries the JSON-encoded request payload.
type ExecuteArgs struct{ Data []byte }

// ExecuteReply carries the JSON-encoded result or an error string.
type ExecuteReply struct {
	Result []byte
	Error  string
}

// ---------------------------------------------------------------------------
// RPCServer — runs inside the plugin subprocess
// ---------------------------------------------------------------------------

// RPCServer wraps a Plugin and exposes it over net/rpc.
// Method signatures follow net/rpc conventions: func(args, *reply) error.
type RPCServer struct{ Impl Plugin }

func (s *RPCServer) Name(_ struct{}, reply *StringReply) error {
	reply.Value = s.Impl.Name()
	return nil
}

func (s *RPCServer) Version(_ struct{}, reply *StringReply) error {
	reply.Value = s.Impl.Version()
	return nil
}

func (s *RPCServer) Initialize(_ struct{}, reply *ErrorReply) error {
	if err := s.Impl.Initialize(); err != nil {
		reply.Error = err.Error()
	}
	return nil
}

func (s *RPCServer) Execute(args ExecuteArgs, reply *ExecuteReply) error {
	var data interface{}
	if err := json.Unmarshal(args.Data, &data); err != nil {
		reply.Error = "decode request: " + err.Error()
		return nil
	}
	result, err := s.Impl.Execute(context.Background(), data)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		reply.Error = "encode result: " + err.Error()
		return nil
	}
	reply.Result = encoded
	return nil
}

func (s *RPCServer) Shutdown(_ struct{}, reply *ErrorReply) error {
	if err := s.Impl.Shutdown(); err != nil {
		reply.Error = err.Error()
	}
	return nil
}

// ---------------------------------------------------------------------------
// RPCClient — runs in the host process, proxies calls to the subprocess
// ---------------------------------------------------------------------------

// RPCClient implements core.Plugin by forwarding every call to the plugin
// subprocess over net/rpc.  Name and Version are cached at connect time.
type RPCClient struct {
	client        *rpc.Client
	cachedName    string
	cachedVersion string
}

func (c *RPCClient) Name() string    { return c.cachedName }
func (c *RPCClient) Version() string { return c.cachedVersion }

func (c *RPCClient) Initialize() error {
	var reply ErrorReply
	if err := c.client.Call("Plugin.Initialize", struct{}{}, &reply); err != nil {
		return fmt.Errorf("rpc Initialize: %w", err)
	}
	if reply.Error != "" {
		return fmt.Errorf("%s", reply.Error)
	}
	return nil
}

func (c *RPCClient) Execute(_ context.Context, data interface{}) (interface{}, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	var reply ExecuteReply
	if err := c.client.Call("Plugin.Execute", ExecuteArgs{Data: encoded}, &reply); err != nil {
		return nil, fmt.Errorf("rpc Execute: %w", err)
	}
	if reply.Error != "" {
		return nil, fmt.Errorf("%s", reply.Error)
	}
	var result interface{}
	if err := json.Unmarshal(reply.Result, &result); err != nil {
		return nil, fmt.Errorf("decode result: %w", err)
	}
	return result, nil
}

func (c *RPCClient) Shutdown() error {
	var reply ErrorReply
	if err := c.client.Call("Plugin.Shutdown", struct{}{}, &reply); err != nil {
		return fmt.Errorf("rpc Shutdown: %w", err)
	}
	if reply.Error != "" {
		return fmt.Errorf("%s", reply.Error)
	}
	return nil
}
