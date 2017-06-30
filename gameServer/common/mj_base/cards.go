package mj_base

import (
	"github.com/lovelly/leaf/util"
)

const (
	IDX_HZMJ = 0
	IDX_ZPMJ = 1
)

var cards = [][]int{
	IDX_HZMJ: []int{ //红中麻将原始麻将数据
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子
		0x35, 0x35, 0x35, 0x35, //红中
	},
	IDX_ZPMJ: []int{ //漳浦麻将数据
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //同子
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //同子
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //同子
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //同子
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //万子
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //万子
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //万子
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //万子
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //索子
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //索子
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //索子
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //索子
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, //番子
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, //番子
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, //番子
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, //番子
		0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, //花子
	},
}

type MJ_CFG struct {
	MaxIdx       int //最多多少种牌
	MaxWeave     int //最多多少种组合
	MaxCount     int //最大手牌数目
	MaxRepertory int //最多存放多少张牌
	HuaIndex     int //花牌开始缩影
	HuaCount     int //花牌数量
}

var cfg = []*MJ_CFG{
	IDX_HZMJ: &MJ_CFG{
		MaxIdx:       42,
		MaxWeave:     5,
		MaxCount:     14,
		MaxRepertory: 112,
		HuaIndex:     0,
		HuaCount:     8,
	},

	IDX_ZPMJ: &MJ_CFG{
		MaxIdx:       42,
		MaxWeave:     5,
		MaxCount:     17,
		MaxRepertory: 144,
		HuaIndex:     8,
		HuaCount:     8,
	},
}

func GetCardByIdx(idx int) []int {
	hzcard := make([]int, len(cards[idx]))
	oldcard := cards[idx]
	util.DeepCopy(&hzcard, &oldcard)
	return hzcard
}

func GetCfg(idx int) *MJ_CFG {
	return cfg[idx]
}
