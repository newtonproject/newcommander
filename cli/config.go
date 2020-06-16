package cli

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
)

const defaultConfigFile = "./config.toml"
const defaultWalletPath = "./wallet/"

type BlockChain int

const (
	UnknownChain BlockChain = iota
	NewChain
	Ethereum
)

func (bc BlockChain) String() string {
	switch bc {
	case NewChain:
		return "NewChain"
	case Ethereum:
		return "Ethereum"
	}

	return "UnknownChain"
}

func (bc BlockChain) Init() {
	InitUnit(bc)

	// default RPC Url
	defaultRPCURL = defaultETHRPCUrl
	if bc == NewChain {
		defaultRPCURL = defaultNEWRPCURL
	}
}

var defaultRPCURL string

const defaultNEWRPCURL = "https://rpc1.newchain.newtonproject.org"
const defaultETHRPCUrl = "https://ethrpc.service.newtonproject.org"

const (
	newchainPublicKey = "c829d38b9fc274c8cb13b239a2b473ec04363167a84f2b4d6666b286f78c92515228bb895ac3802285cde0bac18592efbaffeb1bc14e1da00139b7dbf5248375"
	ethereumPublicKey = "979b7fa28feeb35a4741660a16076f1943202cb72b6af70d327f053e248bab9ba81760f39d0701ef1d8f89cc1fbd2cacba0710a12cd5314d5e0c9021aa3637f9"
)

func (cli *CLI) defaultConfig() {
	viper.BindPFlag("walletPath", cli.rootCmd.PersistentFlags().Lookup("walletPath"))
	viper.BindPFlag("rpcURL", cli.rootCmd.PersistentFlags().Lookup("rpcURL"))

	viper.SetDefault("walletPath", defaultWalletPath)
	viper.SetDefault("rpcURL", defaultRPCURL)
}

func (cli *CLI) setupConfig() error {

	// var ret bool
	var err error

	cli.defaultConfig()

	viper.SetConfigName(defaultConfigFile)
	viper.AddConfigPath(".")
	cfgFile := cli.config
	if cfgFile != "" {
		if _, err = os.Stat(cfgFile); err == nil {
			viper.SetConfigFile(cfgFile)
			err = viper.ReadInConfig()
		} else {
			// The default configuration is enabled.
			// fmt.Println(err)
			err = nil
		}
	} else {
		// The default configuration is enabled.
		err = nil
	}

	if rpcURL := viper.GetString("rpcURL"); rpcURL != "" {
		cli.rpcURL = viper.GetString("rpcURL")
	}
	if walletPath := viper.GetString("walletPath"); walletPath != "" {
		cli.walletPath = viper.GetString("walletPath")
	}

	return cli.setDefaultTransaction()
}

func (cli *CLI) setDefaultTransaction() error {
	if cli.tran == nil {
		cli.tran = new(Transaction)
	}
	fromStr := viper.GetString("pay.from")
	if common.IsHexAddress(fromStr) {
		cli.tran.From = common.HexToAddress(fromStr)
	}
	unit := viper.GetString("pay.unit")
	if stringInSlice(unit, UnitList) {
		cli.tran.Unit = unit
	}
	if password := viper.GetString("pay.password"); password != "" {
		cli.tran.Password = viper.GetString("pay.password")
	}

	return nil
}
