package main

import (
	"mj/common"
	"mj/common/consul"
	"mj/gameServer/Chat"
	"mj/gameServer/center"
	"mj/gameServer/conf"
	"mj/gameServer/db"
	"mj/gameServer/db/model/base"
	"mj/gameServer/gate"
	"mj/gameServer/kindList"
	"mj/gameServer/userHandle"

	"flag"

	"fmt"
	"os"

	"mj/gameServer/http_service"

	"github.com/lovelly/leaf"
	lconf "github.com/lovelly/leaf/conf"
	"github.com/lovelly/leaf/module"
)

var version = 0

var printVersion = flag.Bool("version", false, "print version")

func main() {
	if *printVersion {
		fmt.Println(" version: ", version)
		os.Exit(0)
	}
	Init()
	common.Init()
	http_service.StartHttpServer()
	http_service.StartPrivateServer()
	consul.SetConfig(&conf.ConsulConfig{})
	consul.SetSelfId(conf.ServerName())
	db.InitDB(&conf.DBConfig{})
	base.LoadBaseData()
	kindList.Init()

	modules := []module.Module{center.Module}
	modules = append(modules, gate.Module)
	modules = append(modules, center.Module)
	modules = append(modules, consul.Module)
	modules = append(modules, Chat.Module)
	modules = append(modules, userHandle.UserMgr)
	modules = append(modules, kindList.GetModules()...)
	leaf.Run(modules...)
}

func Init() {
	flag.Parse()
	conf.Init()
	lconf.LogLevel = conf.Server.LogLevel
	lconf.LogPath = conf.Server.LogPath
	lconf.LogFlag = conf.LogFlag
	lconf.ConsolePort = conf.Server.ConsolePort
	lconf.ProfilePath = conf.Server.ProfilePath
	lconf.ListenAddr = conf.Server.ListenAddr
	lconf.ServerName = conf.ServerName()
	lconf.ConnAddrs = conf.Server.ConnAddrs
	lconf.PendingWriteNum = conf.Server.PendingWriteNum
	lconf.HeartBeatInterval = conf.HeartBeatInterval
	leaf.InitLog()
}
