package cli

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	prompt2 "github.com/ethereum/go-ethereum/console/prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (cli *CLI) buildInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "init",
		Short:                 "Initialize config file",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {

			fmt.Println("Initialize config file")

			prompt := fmt.Sprintf("Enter file in which to save (%s): ", defaultConfigFile)
			configPath, err := prompt2.Stdin.PromptInput(prompt)
			if err != nil {
				fmt.Println("PromptInput err:", err)
			}
			if configPath == "" {
				configPath = defaultConfigFile
			}
			cli.config = configPath

			walletPathV := viper.GetString("walletPath")
			prompt = fmt.Sprintf("Enter the wallet storage directory (%s): ", walletPathV)
			walletPath, err := prompt2.Stdin.PromptInput(prompt)
			if err != nil {
				fmt.Println("PromptInput err:", err)
			}
			if walletPath == "" {
				walletPath = walletPathV
			}
			viper.Set("walletPath", walletPath)

			rpcURLV := viper.GetString("rpcURL")
			prompt = fmt.Sprintf("Enter %s json rpc or ipc url (%s): ", cli.blockchain.String(), rpcURLV)
			rpcURL, err := prompt2.Stdin.PromptInput(prompt)
			if err != nil {
				fmt.Println("PromptInput err:", err)
			}
			if rpcURL == "" {
				rpcURL = rpcURLV
			}
			viper.Set("rpcURL", rpcURL)

			prompt = fmt.Sprintf("Create a new account or not: [Y/n] ")
			createNewAddress, err := prompt2.Stdin.PromptInput(prompt)
			if err != nil {
				fmt.Println("PromptInput err:", err)
			}
			if len(createNewAddress) <= 0 {
				createNewAddress = "Y"
			}
			if strings.ToUpper(createNewAddress[:1]) == "Y" {
				wallet := keystore.NewKeyStore(walletPath,
					keystore.StandardScryptN, keystore.StandardScryptP)

				walletPassword, err := getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true)
				if err == nil {
					account, err := wallet.NewAccount(walletPassword)
					if err == nil {
						fromAddress := account.Address.String()
						fmt.Println(fromAddress)
						viper.Set("pay.from", fromAddress)
					} else {
						fmt.Println("Account error:", err)
						fmt.Println("Just create your account later.")
					}
				} else {
					fmt.Println("Error: ", err)
					fmt.Println("Just create your account later.")
				}
			}

			err = viper.WriteConfigAs(configPath)
			if err != nil {
				fmt.Println("WriteConfig:", err)
			} else {
				fmt.Println("Your configuration has been saved in ", configPath)
			}
		},
	}

	return cmd
}
