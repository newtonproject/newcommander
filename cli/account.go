package cli

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	prompt2 "github.com/ethereum/go-ethereum/console/prompt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (cli *CLI) buildAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [new|list|update]",
		Short: fmt.Sprintf("Manage %s accounts", cli.blockchain.String()),
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			return
		},
	}

	cmd.AddCommand(cli.buildAccountNewCmd())
	cmd.AddCommand(cli.buildAccountListCmd())
	cmd.AddCommand(cli.buildAccountUpdateCmd())
	cmd.AddCommand(cli.buildAccountImportCmd())
	cmd.AddCommand(cli.buildAccountExportCmd())

	if cli.blockchain == NewChain {
		cmd.AddCommand(cli.buildAccountConvertCmd())
	}

	return cmd
}

func (cli *CLI) buildAccountNewCmd() *cobra.Command {
	accountNewCmd := &cobra.Command{
		Use:                   "new [-n number] [--faucet] [-s] [-l]",
		Short:                 "create a new account",
		Args:                  cobra.MinimumNArgs(0),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flags().Changed("light") {
				light, _ := cmd.Flags().GetBool("light")
				standard, _ := cmd.Flags().GetBool("standard")
				if light && !standard {
					cli.wallet = keystore.NewKeyStore(cli.walletPath,
						keystore.LightScryptN, keystore.LightScryptP)
				}
			}

			numOfNew, err := cmd.Flags().GetInt("numOfNew")
			if err != nil {
				numOfNew = viper.GetInt("account.numOfNew")
			}

			faucet, _ := cmd.Flags().GetBool("faucet")

			aList, err := cli.createAccount(numOfNew)
			if err != nil {
				fmt.Println(err)
				return
			}

			for _, a := range aList {
				fmt.Println(a.String())
				if faucet {
					getFaucet(cli.rpcURL, a.String())
				}
			}

		},
	}

	accountNewCmd.Flags().IntP("numOfNew", "n", 1, "number of the new account")
	accountNewCmd.Flags().Bool("faucet", false, "get faucet for new account")
	accountNewCmd.Flags().BoolP("standard", "s", false, "use the standard scrypt for keystore")
	accountNewCmd.Flags().BoolP("light", "l", false, "use the light scrypt for keystore")
	return accountNewCmd
}

func (cli *CLI) createAccount(numOfNew int) ([]common.Address, error) {
	if cli.wallet == nil {
		cli.wallet = keystore.NewKeyStore(cli.walletPath,
			keystore.StandardScryptN, keystore.StandardScryptP)
	}

	walletPassword, err := getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true)
	if err != nil {
		return nil, err
	}

	if numOfNew <= 0 {
		fmt.Printf("number[%d] of new account less then 1\n", numOfNew)
		numOfNew = 1
	}

	aList := make([]common.Address, 0)
	for i := 0; i < numOfNew; i++ {
		account, err := cli.wallet.NewAccount(walletPassword)
		if err != nil {
			return nil, fmt.Errorf("account error: %v", err)
		}

		aList = append(aList, account.Address)
	}

	return aList, nil
}

func (cli *CLI) buildAccountListCmd() *cobra.Command {
	accountListCmd := &cobra.Command{
		Use:                   "list",
		Short:                 "list all accounts in the wallet path",
		Args:                  cobra.MinimumNArgs(0),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cli.openWallet(true); err != nil {
				fmt.Println(err)
				return
			}

			pubkey, _ := cmd.Flags().GetBool("pubkey")
			password := ""

			for _, account := range cli.wallet.Accounts() {
				if pubkey {
					keyJSON, err := cli.wallet.Export(account, password, password)
					if err != nil {
						prompt := fmt.Sprintf("Unlocking account %s", account.Address.String())
						password, err = getPassPhrase(prompt, false)
						if err != nil {
							fmt.Println(err)
							return
						}
					}
					keyJSON, err = cli.wallet.Export(account, password, password)
					if err != nil {
						fmt.Println(err)
						return
					}

					key, err := keystore.DecryptKey(keyJSON, password)
					if err != nil {
						fmt.Println(err)
						return
					}

					pub := key.PrivateKey.PublicKey
					fmt.Println(account.Address.Hex(), hex.EncodeToString(crypto.FromECDSAPub(&pub)[1:]), hex.EncodeToString(crypto.Keccak256(crypto.FromECDSAPub(&pub)[1:])))

				} else {
					fmt.Println(account.Address.Hex())
				}
			}
		},
	}
	accountListCmd.Flags().BoolP("pubkey", "s", false, "Show pub key and keccak256 of pub key")

	return accountListCmd
}

func (cli *CLI) buildAccountConvertCmd() *cobra.Command {
	accountListCmd := &cobra.Command{
		Use:                   "convert",
		Short:                 "convert address to NewChainAddress",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Run: func(cmd *cobra.Command, args []string) {

			chainID := DefaultChainID
			err := cli.BuildClient()
			if err != nil {
				fmt.Printf("Error: build client error(%v), use chainID as %s\n", err, chainID.String())
			} else {
				networkID, err := cli.client.NetworkID(context.Background())
				if err != nil {
					fmt.Printf("Error: get chainID error(%v), use chainID as %s\n", err, chainID.String())
				} else {
					chainID = networkID
				}
			}

			for _, addressStr := range args {
				if common.IsHexAddress(addressStr) {
					address := common.HexToAddress(addressStr)
					fmt.Println(address.String(), addressToNew(chainID.Bytes(), address))
					continue
				}

				address, err := newToAddress(chainID.Bytes(), addressStr)
				if err != nil {
					fmt.Println(err, addressStr)
					continue
				}
				fmt.Println(address.String(), addressStr)
			}

		},
	}

	return accountListCmd
}

func (cli *CLI) buildAccountUpdateCmd() *cobra.Command {
	accountNewCmd := &cobra.Command{
		Use:                   "update <address> [-s]",
		Short:                 "Update an existing account",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			standard, _ := cmd.Flags().GetBool("standard")
			if standard {
				cli.wallet = keystore.NewKeyStore(cli.walletPath,
					keystore.StandardScryptN, keystore.StandardScryptP)
			} else {
				cli.wallet = keystore.NewKeyStore(cli.walletPath,
					keystore.LightScryptN, keystore.LightScryptP)
			}

			addrStr := args[0]
			var address common.Address
			if !common.IsHexAddress(addrStr) {
				fmt.Println("Error: No accounts specified to update")
				return
			}
			address = common.HexToAddress(addrStr)
			account := accounts.Account{Address: address}

			if account.Address == (common.Address{}) {
				fmt.Println("Error: ", errRequiredFromAddress)
				return
			}
			if _, err := cli.wallet.Find(account); err != nil {
				fmt.Printf("Error: %v (%s)\n", err, account.Address.String())
				return
			}

			var err error
			walletPassword := cli.tran.Password
			var trials int
			for trials = 0; trials < 3; trials++ {
				prompt := fmt.Sprintf("Unlocking account %s | Attempt %d/%d", account.Address.String(), trials+1, 3)
				if walletPassword == "" {
					walletPassword, _ = getPassPhrase(prompt, false)
				} else {
					fmt.Println(prompt, "\nUse the the password has set")
				}
				err = cli.wallet.Unlock(account, walletPassword)
				if err == nil {
					break
				}
				walletPassword = ""
			}

			if trials >= 3 {
				if err != nil {
					fmt.Println("Error: unlock account error: ", err)
					return
				}
				fmt.Printf("Error: failed to unlock account %s (%v)\n", account.Address.String(), err)
				return
			}

			newWalletPassword, err := getPassPhrase("Please give a new password. Do not forget this password.", true)
			if err != nil {
				fmt.Println("Error: getPassPhrase error: ", err)
				return
			}

			if err := cli.wallet.Update(account, walletPassword, newWalletPassword); err != nil {
				fmt.Println("Error: udpate account error: ", err)
				return
			}

			fmt.Println("Successfully updated the account: ", account.Address.String())
		},
	}

	accountNewCmd.Flags().BoolP("standard", "s", false, "use the standard scrypt for keystore")
	return accountNewCmd
}

func addressToNew(chainID []byte, address common.Address) string {
	input := append(chainID, address.Bytes()...)
	return "NEW" + base58.CheckEncode(input, 0)
}

func newToAddress(chainID []byte, newAddress string) (common.Address, error) {
	if newAddress[:3] != "NEW" {
		return common.Address{}, errors.New("not NEW address")
	}

	decoded, version, err := base58.CheckDecode(newAddress[3:])
	if err != nil {
		return common.Address{}, err
	}
	if version != 0 {
		return common.Address{}, errors.New("illegal version")
	}
	if len(decoded) < 20 {
		return common.Address{}, errors.New("illegal decoded length")
	}
	if !bytes.Equal(decoded[:len(decoded)-20], chainID) {
		return common.Address{}, errors.New("illegal ChainID")
	}

	address := common.BytesToAddress(decoded[len(decoded)-20:])

	return address, nil
}

func (cli *CLI) buildAccountImportCmd() *cobra.Command {
	accountListCmd := &cobra.Command{
		Use:                   "import",
		Short:                 "import hex private key to wallet",
		Args:                  cobra.MinimumNArgs(0),
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Run: func(cmd *cobra.Command, args []string) {
			hexkey, err := prompt2.Stdin.PromptPassword("Enter private key: ")
			if err != nil {
				fmt.Println(err)
				return
			}
			hexkeylen := len(hexkey)
			if hexkeylen > 1 {
				if hexkey[0:2] == "0x" || hexkey[0:2] == "0X" {
					hexkey = hexkey[2:]
				}
			}

			pkey, err := crypto.HexToECDSA(hexkey)
			if err != nil {
				fmt.Println(err)
				return
			}

			if err := cli.openWallet(false); err != nil {
				fmt.Println(err)
				return
			}

			walletPassword, err := getPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true)
			if err != nil {
				fmt.Println("Error: ", err)
				return
			}

			a, err := cli.wallet.ImportECDSA(pkey, walletPassword)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(a.Address.String())

		},
	}

	return accountListCmd
}

func (cli *CLI) buildAccountExportCmd() *cobra.Command {
	accountListCmd := &cobra.Command{
		Use:                   "export <hexAddress>",
		Short:                 "export hex private key of the specified address",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Run: func(cmd *cobra.Command, args []string) {

			hexAddress := args[0]
			if !common.IsHexAddress(hexAddress) {
				fmt.Printf("Error: %s is not valid hex-encoded address\n", hexAddress)
				return
			}

			address := common.HexToAddress(hexAddress)
			account := accounts.Account{Address: address}

			if err := cli.openWallet(true); err != nil {
				fmt.Println(err)
				return
			}

			if !cli.wallet.HasAddress(address) {
				fmt.Println("The given address is not present")
				return
			}

			var err error
			walletPassword := cli.tran.Password
			if walletPassword == "" {
				prompt := fmt.Sprintf("Unlocking account %s", account.Address.String())
				walletPassword, _ = getPassPhrase(prompt, false)
			}
			keyJSON, err := cli.wallet.Export(account, walletPassword, walletPassword)
			if err != nil {
				fmt.Println(err)
				return
			}

			key, err := keystore.DecryptKey(keyJSON, walletPassword)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(common.Bytes2Hex(key.PrivateKey.D.Bytes()))

		},
	}

	return accountListCmd
}
