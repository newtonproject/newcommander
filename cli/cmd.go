package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (cli *CLI) buildRootCmd() {
	if cli.rootCmd != nil {
		cli.rootCmd.ResetFlags()
		cli.rootCmd.ResetCommands()
	}

	short := fmt.Sprintf("%s is a commandline client for the %s blockchain", cli.Name, cli.blockchain.String())
	rootCmd := &cobra.Command{
		Use:              cli.Name, // newcommander
		Short:            short,
		Run:              cli.help,
		PersistentPreRun: cli.setup,
	}
	cli.rootCmd = rootCmd

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cli.config, "config", "c", defaultConfigFile, "The `path` to config file")
	rootCmd.PersistentFlags().StringP("walletPath", "w", defaultWalletPath, "Wallet storage `directory`")
	rootCmd.PersistentFlags().StringP("rpcURL", "i", defaultRPCURL, fmt.Sprintf("%s json rpc or ipc `url`", cli.blockchain.String()))

	// Basic commands
	rootCmd.AddCommand(cli.buildVersionCmd()) // version
	rootCmd.AddCommand(cli.buildInitCmd())    // init
	rootCmd.AddCommand(cli.buildAccountCmd()) // account
	rootCmd.AddCommand(cli.buildBalanceCmd()) // balance

	// Core commands
	rootCmd.AddCommand(cli.buildPayCmd()) // pay

	// Aux commands
	rootCmd.AddCommand(cli.buildFaucetCmd()) // faucet

	// Alias commands

	// offline tools
	rootCmd.AddCommand(cli.buildBuildCmd())     // build tx
	rootCmd.AddCommand(cli.buildSignCmd())      // sign tx
	rootCmd.AddCommand(cli.buildBroadcastCmd()) // submit/broadcast

	// rpc
	rootCmd.AddCommand(cli.buildRPCCmd()) // rpc

	// batch pay
	rootCmd.AddCommand(cli.buildBatchPayCmd()) // batch

	// tools
	rootCmd.AddCommand(cli.buildDecodeCmd()) // decode
	rootCmd.AddCommand(cli.buildBlockCmd())  // block
	rootCmd.AddCommand(cli.buildTraceCmd())  // trace
}
