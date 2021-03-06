package room

import (
	"mj/common/msg/mj_hz_msg"
	"mj/gameServer/common/mj/mj_base"
	"mj/gameServer/user"

	"github.com/lovelly/leaf/gate"
)

func NewHZEntry(kindID, ServerID int) *hz_entry {
	e := new(hz_entry)
	e.Mj_base = mj_base.NewMJBase(kindID, ServerID)
	return e
}

type hz_entry struct {
	*mj_base.Mj_base
}

func (e *hz_entry) GetDataMgr() *hz_data{
	return e.DataMgr.(*hz_data)
}

func (e *hz_entry) ZhaMa(args []interface{}) {
	//recvMsg := args[0].(*mj_hz_msg.C2G_HZMJ_ZhaMa)
	retMsg := &mj_hz_msg.G2C_HZMJ_ZhuaHua{}
	agent := args[1].(gate.Agent)
	u := agent.UserData().(*user.User)

	winUser := []int{u.ChairId}
	ZhongHua, BuZhong := e.GetDataMgr().OnZhuaHua(winUser)
	for card := range ZhongHua[0] {
		retMsg.ZhongHua = append(retMsg.ZhongHua, card)
	}
	for card := range BuZhong {
		retMsg.BuZhong = append(retMsg.BuZhong, card)
	}
	u.WriteMsg(retMsg)
}
