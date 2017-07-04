package internal

import (
	"errors"
	"mj/hallServer/db"
	"mj/hallServer/user"
)

type Order struct {
	OnLineID    int `db:"OnLineID"`    //
	PayAmount   int `db:"PayAmount"`   // 价格
	UserID      int `db:"UserID"`      // uid
	GoodsID     int `db:"GoodsID"`     // 物品id
	OrderStatus int `db:"OrderStatus"` //状态
}

func (m *UserModule) GetUser(args []interface{}) (interface{}, error) {
	u, ok := m.a.UserData().(*user.User)
	if !ok {
		return nil, errors.New("not foud user Data at GetUser")
	}
	return u, nil
}

func GetOrder(uid int) []*Order {
	obj := make([]*Order, 0)
	err := db.AccountDB.Select(&obj, "select OnLineID,PayAmount, UserID, GoodsID, OrderStatus FROM `order` where UserID=?", uid)
	if err != nil {
		return nil
	}

	if len(obj) < 1 {
		return nil
	}
	return obj
}
