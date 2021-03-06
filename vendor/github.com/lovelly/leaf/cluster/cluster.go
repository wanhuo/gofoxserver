package cluster

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/lovelly/leaf/chanrpc"
	"github.com/lovelly/leaf/conf"
	"github.com/lovelly/leaf/log"
	"github.com/lovelly/leaf/network"
	lgob "github.com/lovelly/leaf/network/gob"
)

const (
	NeedWaitRequestTimes = 5
)

var (
	closing      bool
	closeSig     = make(chan bool, 1)
	wg           sync.WaitGroup
	server       *network.TCPServer
	clientsMutex sync.Mutex
	clients      = map[string]*network.TCPClient{}
	agentsMutex  sync.RWMutex
	agents       = map[string]*Agent{}
	AgentChanRPC *chanrpc.Server
)

func Init() {
	if conf.ListenAddr != "" {
		server = new(network.TCPServer)
		server.Addr = conf.ListenAddr
		server.MaxConnNum = int(math.MaxInt32)
		server.PendingWriteNum = conf.PendingWriteNum
		server.LenMsgLen = 4
		server.MaxMsgLen = math.MaxUint32
		server.NewAgent = newAgent

		server.Start()
	}


}

func AddClient(serverName, addr string) {
	log.Debug("at cluster AddClient %s, %s", serverName, addr)
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	_removeClient(serverName)

	client := new(network.TCPClient)
	client.Addr = addr
	client.ConnNum = 1
	client.ConnectInterval = 3 * time.Second
	client.PendingWriteNum = conf.PendingWriteNum
	client.LenMsgLen = 4
	client.MaxMsgLen = math.MaxUint32
	client.NewAgent = newAgent
	client.AutoReconnect = true

	client.Start()
	clients[serverName] = client
}

func _removeClient(serverName string) {
	client, ok := clients[serverName]
	if ok {
		log.Debug("at cluster _removeClient %s", serverName)
		client.Close()
		delete(clients, serverName)
	}
}

func RemoveClient(serverName string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	_removeClient(serverName)
}

func addAgent(serverName string, agent *Agent) {
	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	_removeAgent(serverName)

	agent.ServerName = serverName
	agents[agent.ServerName] = agent
	log.Release("%v server is online", serverName)

	if AgentChanRPC != nil {
		AgentChanRPC.Go("NewServerAgent", serverName, agent)
	}
}

func _removeAgent(serverName string) {
	agent, ok := agents[serverName]
	if ok {
		log.Debug("remove svr name ==== == %v", serverName)
		delete(agents, serverName)
		agent.Destroy()
		log.Release("%v server is offline", serverName)

		if AgentChanRPC != nil {
			AgentChanRPC.Go("CloseServerAgent", serverName, agent)
		}
	}
}

func removeAgent(serverName string) {
	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	_removeAgent(serverName)
}

func Destroy() {
	closing = true
	waitRequestTimes := 0
	for {
		time.Sleep(time.Second)

		requestCount := GetRequestCount()
		if requestCount == 0 {
			waitRequestTimes += 1
			if waitRequestTimes >= NeedWaitRequestTimes {
				break
			} else {
				log.Release("wait request count down %v", NeedWaitRequestTimes-waitRequestTimes)
			}
		} else {
			waitRequestTimes = 0
			log.Release("has %v request", requestCount)
		}
	}

	closeSig <- true
	wg.Wait()

	if server != nil {
		server.Close()
	}

	clientsMutex.Lock()
	for _, client := range clients {
		client.Close()
	}
	clientsMutex.Unlock()
}

type Agent struct {
	ServerName         string
	conn               *network.TCPConn
	userData           interface{}
	heartBeatWaitTimes int32

	encMutex sync.Mutex
	encoder  *lgob.Encoder
	decoder  *lgob.Decoder

	sync.Mutex
	requestID  uint32
	requestMap map[uint32]*RequestInfo
}

func newAgent(conn *network.TCPConn) network.Agent {
	a := new(Agent)
	a.conn = conn
	a.requestMap = make(map[uint32]*RequestInfo)

	a.encoder = lgob.NewEncoder()
	a.decoder = lgob.NewDecoder()

	msg := &S2S_NotifyServerName{ServerName: conf.ServerName}
	a.WriteMsg(msg)
	return a
}

func (a *Agent) GetRequestCount() int {
	a.Lock()
	defer a.Unlock()
	return len(a.requestMap)
}

func (a *Agent) registerRequest(request *RequestInfo) uint32 {
	a.Lock()
	defer a.Unlock()

	reqID := a.requestID
	a.requestMap[reqID] = request
	a.requestID += 1
	return reqID
}

func (a *Agent) popRequest(requestID uint32) *RequestInfo {
	a.Lock()
	defer a.Unlock()

	request, ok := a.requestMap[requestID]
	if ok {
		delete(a.requestMap, requestID)
		return request
	} else {
		return nil
	}
}

func (a *Agent) clearRequest(err error) {
	a.Lock()
	defer a.Unlock()

	for _, request := range a.requestMap {
		ret := &chanrpc.RetInfo{Err: err, Cb: request.cb}
		request.chanRet <- ret
	}
	a.requestMap = make(map[uint32]*RequestInfo)
}

func (a *Agent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			if err != io.EOF {
				log.Error("read message: %v", err)
			}
			break
		}

		if Processor != nil {
			msg, err := Processor.Unmarshal(a.decoder, data)
			if err != nil {
				log.Error("unmarshal message error: %v", err)
				continue
			}
			log.Debug("cluster IN = %v", msg)
			err = Processor.Route(msg, a)
			if err != nil {
				log.Error("route message error: %v", err)
				break
			}
		}
	}
}

func (a *Agent) OnClose() {
	removeAgent(a.ServerName)
	a.clearRequest(fmt.Errorf("%v server is offline", a.ServerName))
}

func (a *Agent) WriteMsg(msg interface{}) {

	if Processor != nil {
		log.Debug("cluster OUT = %v", msg)
		a.encMutex.Lock()
		data, err := Processor.Marshal(a.encoder, msg)
		a.encMutex.Unlock()
		if err != nil {
			log.Error("marshal message %v error: %v", reflect.TypeOf(msg), err)
			return
		}
		if len(data) < 1 {
			log.Error("cluster WriteMsg msg len == 0")
			return
		}
		err = a.conn.WriteMsg(data...)
		if err != nil {
			log.Error("write message %v error: %v", reflect.TypeOf(msg), err)
		}
	}

}

func (a *Agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *Agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *Agent) Close() {
	a.conn.Close()
}

func (a *Agent) Destroy() {
	a.conn.Destroy()
}

func (a *Agent) UserData() interface{} {
	return a.userData
}

func (a *Agent) SetUserData(data interface{}) {
	a.userData = data
}

func (a *Agent) Go(id interface{}, args ...interface{}) {
	msg := &S2S_RequestMsg{MsgID: id, CallType: callNotForResult, Args: args}
	a.WriteMsg(msg)
}

//timeOutCall 会丢弃执行结果
func (a *Agent) TimeOutCall1(id interface{}, t time.Duration, args ...interface{}) (interface{}, error) {
	chanSyncRet := make(chan *chanrpc.RetInfo, 1)

	request := &RequestInfo{chanRet: chanSyncRet}
	requestID := a.registerRequest(request)
	msg := &S2S_RequestMsg{RequestID: requestID, MsgID: id, CallType: callForResult, Args: args}
	a.WriteMsg(msg)
	select {
	case ri := <-chanSyncRet:
		return ri.Ret, ri.Err
	case <-time.After(time.Second * t):
		a.popRequest(requestID)
		return nil, errors.New(fmt.Sprintf("time out at TimeOutCall1 function: %v", id))
	}
}

func (a *Agent) Call0(id interface{}, args ...interface{}) error {
	chanSyncRet := make(chan *chanrpc.RetInfo, 1)

	request := &RequestInfo{chanRet: chanSyncRet}
	requestID := a.registerRequest(request)
	msg := &S2S_RequestMsg{RequestID: requestID, MsgID: id, CallType: callForResult, Args: args}
	a.WriteMsg(msg)

	ri := <-chanSyncRet
	return ri.Err
}

func (a *Agent) Call1(id interface{}, args ...interface{}) (interface{}, error) {
	chanSyncRet := make(chan *chanrpc.RetInfo, 1)

	request := &RequestInfo{chanRet: chanSyncRet}
	requestID := a.registerRequest(request)
	msg := &S2S_RequestMsg{RequestID: requestID, MsgID: id, CallType: callForResult, Args: args}
	a.WriteMsg(msg)

	ri := <-chanSyncRet
	return ri.Ret, ri.Err
}

func (a *Agent) CallN(id interface{}, args ...interface{}) ([]interface{}, error) {
	chanSyncRet := make(chan *chanrpc.RetInfo, 1)

	request := &RequestInfo{chanRet: chanSyncRet}
	requestID := a.registerRequest(request)
	msg := &S2S_RequestMsg{RequestID: requestID, MsgID: id, CallType: callForResult, Args: args}
	a.WriteMsg(msg)

	ri := <-chanSyncRet
	return chanrpc.Assert(ri.Ret), ri.Err
}

func (a *Agent) AsynCall(chanAsynRet chan *chanrpc.RetInfo, id interface{}, args ...interface{}) {
	if len(args) < 1 {
		panic(fmt.Sprintf("%v asyn call of callback function not found", id))
	}

	lastIndex := len(args) - 1
	cb := args[lastIndex]
	args = args[:lastIndex]

	var callType uint8
	switch cb.(type) {
	case func(error):
		callType = callForResult
	case func(interface{}, error):
		callType = callForResult
	case func([]interface{}, error):
		callType = callForResult
	default:
		panic(fmt.Sprintf("%v asyn call definition of callback function is invalid", id))
	}

	request := &RequestInfo{cb: cb, chanRet: chanAsynRet}
	requestID := a.registerRequest(request)
	msg := &S2S_RequestMsg{RequestID: requestID, MsgID: id, CallType: callType, Args: args}
	a.WriteMsg(msg)
}
