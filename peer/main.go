/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strings"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/config"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/peer/chaincode"
	"github.com/hyperledger/fabric/peer/channel"
	"github.com/hyperledger/fabric/peer/clilogging"
	"github.com/hyperledger/fabric/peer/common"
	"github.com/hyperledger/fabric/peer/node"
	"github.com/hyperledger/fabric/peer/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logger = flogging.MustGetLogger("main")
var logOutput = os.Stderr

// Constants go here.
const cmdRoot = "core"

// The main command describes the service and
// defaults to printing the help message.
// 主命令，默认打印帮助
var mainCmd = &cobra.Command{
	Use: "peer",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// runE之前执行
		// check for --logging-level pflag first, which should override all other
		// log settings. if --logging-level is not set, use CORE_LOGGING_LEVEL
		// (environment variable takes priority; otherwise, the value set in
		// core.yaml)
		var loggingSpec string
		if viper.GetString("logging_level") != "" {
			loggingSpec = viper.GetString("logging_level")
		} else {
			loggingSpec = viper.GetString("logging.level")
		}
		flogging.InitFromSpec(loggingSpec)

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Print(version.GetInfo())
		} else {
			cmd.HelpFunc()(cmd, args)
		}
	},
}

// Peer command version flag
var versionFlag bool

func main() {
	// For environment variables.
	//设置环境变量前缀，peer为core，环境变量以CORE_开头。
	viper.SetEnvPrefix(cmdRoot)
	viper.AutomaticEnv() // 自动前缀
	replacer := strings.NewReplacer(".", "_") // 字符替换
	viper.SetEnvKeyReplacer(replacer)

	// Define command-line flags that are valid for all peer commands and
	// subcommands.
	mainFlags := mainCmd.PersistentFlags()
	mainFlags.BoolVarP(&versionFlag, "version", "v", false, "Display current version of fabric peer server")

	mainFlags.String("logging-level", "", "Default logging level and overrides, see core.yaml for full syntax")
	// 绑定命令行logging-level值
	viper.BindPFlag("logging_level", mainFlags.Lookup("logging-level"))

	// 添加子命令
	mainCmd.AddCommand(version.Cmd())
	mainCmd.AddCommand(node.Cmd())
	mainCmd.AddCommand(chaincode.Cmd(nil))
	mainCmd.AddCommand(clilogging.Cmd(nil))
	mainCmd.AddCommand(channel.Cmd(nil))

	// 初始化viper组件，添加config路径
	//addConfigPath(v, altPath)
	//addConfigPath(v, "./")
	//AddDevConfigPath(v)
	//AddDevConfigPath("/etc/hyperledger/fabric")
	// 执行viper.ReadInConfig()
	err := common.InitConfig(cmdRoot)
	if err != nil { // Handle errors reading the config file
		logger.Errorf("Fatal error when initializing %s config : %s", cmdRoot, err)
		os.Exit(1)
	}
    // 获取peer变量，PEER_GOMAXPROCS，并设置GOMAXPROCS
	runtime.GOMAXPROCS(viper.GetInt("peer.gomaxprocs"))

	// 设置日志格式
	flogging.InitBackend(flogging.SetFormat(viper.GetString("logging.format")), logOutput)

	// 初始化msp
	var mspMgrConfigDir = config.GetPath("peer.mspConfigPath")
	// 获取mspid
	var mspID = viper.GetString("peer.localMspId")
	// 获取msptype，默认fabric
	var mspType = viper.GetString("peer.localMspType")
	if mspType == "" {
		mspType = msp.ProviderTypeToString(msp.FABRIC)
	}
	err = common.InitCrypto(mspMgrConfigDir, mspID, mspType)
	if err != nil { // Handle errors reading the config file
		logger.Errorf("Cannot run peer because %s", err.Error())
		os.Exit(1)
	}
	// 执行具体命令
	if mainCmd.Execute() != nil {
		os.Exit(1)
	}
	logger.Info("Exiting.....")
}
