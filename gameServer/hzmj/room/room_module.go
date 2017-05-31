package room

import (
	"github.com/lovelly/leaf/module"
	"mj/gameServer/base"
	"github.com/lovelly/leaf/chanrpc"
	"github.com/lovelly/leaf/log"
	"time"
	"mj/gameServer/conf"
	"mj/gameServer/common"
	"fmt"
	"strconv"
	"sync"
	"mj/common/msg"
	tbase "mj/gameServer/db/model/base"
)

var (
	idLock sync.RWMutex
	IncId = 0
)

func getId()int {
	idLock.Lock()
	defer  idLock.Unlock()
	IncId ++
	return IncId
}

func NewRoom(mgrCh* chanrpc.Server, param *msg.C2G_CreateTable, t *tbase.GameServiceOption) *Room {
	skeleton := base.NewSkeleton()
	Room := new(Room)
	Room.Skeleton = skeleton
	Room.ChanRPC= skeleton.ChanRPCServer
	Room.mgrCh =mgrCh
	Room.RoomInfo = common.NewRoomInfo()
	Room.id = getId()
	Room.Kind = t.KindID
	Room.ServerId = t.ServerID
	Room.name = fmt.Sprintf( strconv.Itoa(common.KIND_TYPE_HZMJ) +"_%v", Room.id)
	Room.CloseSig = make(chan bool, 1)
	RegisterHandler(Room)
	Room.OnInit()
	go Room.run()
	log.Debug("new room ok .... ")
	return Room
}

//吧room 当一张桌子理解
type Room struct {
	*common.RoomInfo
	*module.Skeleton
	ChanRPC *chanrpc.Server //接受客户端消息的chan
	mgrCh* chanrpc.Server  //管理类的chan 例如红中麻将 就是红中麻将module的 ChanRPC
	name          string
	CloseSig  chan bool
	wg       sync.WaitGroup
	id 			int
	Kind 		int
	ServerId    int
}

func (r *Room)run(){
	r.wg.Add(1)
	log.Debug("room Room start run Name:%s", r.name)
	r.Run(r.CloseSig)
	log.Debug("room Room End run Name:%s", r.name)
	r.wg.Done()
}

func  (r *Room) Destroy(){
	r.CloseSig <- true
	r.wg.Wait()
	r.OnDestroy()
	log.Debug("room Room Destroy ok,  Name:%s", r.name)
}


////////////////// 上面run 和 Destroy 请勿随意修改 //////  下面函数自由操作
func (r *Room) OnInit() {
	r.Skeleton.AfterFunc(time.Duration(conf.DestroyRoomInterval/10), r.checkDestroyRoom)
}

func (r *Room) OnDestroy() {

}

func (r *Room) GetRoomId() int{
	return r.id
}

//这里添加定时操作
func (r *Room) checkDestroyRoom() {
	nowTime := time.Now().Unix()
	if r.CheckDestroy(nowTime) {
		r.Destroy()
		return
	}

	r.Skeleton.AfterFunc(time.Duration(conf.DestroyRoomInterval/10), r.checkDestroyRoom)
}

func (r *Room) GetChanRPC() *chanrpc.Server {
	return r.ChanRPC
}



