package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	prompt2 "github.com/ethereum/go-ethereum/console/prompt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/spf13/cobra"
)

func (cli *CLI) buildBuildCmd() *cobra.Command {
	buildTxCmd := &cobra.Command{
		Use:                   "build [--out outfile]",
		Short:                 "Build transaction",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var inStr string

			if cmd.Flags().Changed("in") {
				inStr, err = cmd.Flags().GetString("in")
				if err != nil {
					fmt.Println(err)
					return
				}
				if err := cli.applyTxFile(inStr); err != nil {
					fmt.Printf("Error apply infile(%s): %v\n", inStr, err)
					return
				}
			}

			offline, _ := cmd.Flags().GetBool("offline")

			if cmd.Flags().Changed("noguide") {
				if ok, _ := cmd.Flags().GetBool("noguide"); !ok {
					fmt.Println("flag noguide changed but is false")
					return
				}
			} else {
				if err := cli.applyTxGuide(offline); err != nil {
					fmt.Println(err)
					return
				}
			}

			// update nonce, gasPrice, gasLimit, networkID from node
			if !offline {
				fmt.Println("Updating nonce, gasPrice, gasLimit and networkID from node...")
				if err := cli.updateFromNode(); err != nil {
					fmt.Println(err)
					return
				}
			}

			tByte, err := cli.tran.MarshalJSON()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Transaction details are as follows:")
			fmt.Println(string(tByte))

			var outStr string
			defaultOutStr := time.Now().Format("20060102150405") + ".tx" // bitcoin 2009-01-03 18:15:05
			if cmd.Flags().Changed("out") {
				outStr, err = cmd.Flags().GetString("out")
			} else {
				prompt := fmt.Sprintf("Enter file to save transaction (default: %s): ", defaultOutStr)
				outStr, err = prompt2.Stdin.PromptInput(prompt)

			}
			if err != nil {
				fmt.Println("Error:", err)
			}
			if outStr == "" {
				outStr = defaultOutStr
			}
			if err := cli.saveTranToFile(outStr); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Successfully save transaction to file", outStr)
			}

			if sign, _ := cmd.Flags().GetBool("sign"); !sign && !offline {
				return
			}

			if inStr == "" {
				inStr = "tx"
			}
			cli.signTxAndSave(inStr + ".sign")
		},
	}

	buildTxCmd.Flags().String("out", "", "file `path` to save built transaction")
	buildTxCmd.Flags().String("in", "", "file `path` to load transaction to be built")
	buildTxCmd.Flags().Bool("noguide", false, "disable guide to build transaction")
	buildTxCmd.Flags().Bool("sign", false, "sign transaction after build")
	buildTxCmd.Flags().Bool("offline", false, "build offline transaction")

	return buildTxCmd
}

func (cli *CLI) buildSignCmd() *cobra.Command {
	signTxCmd := &cobra.Command{
		Use:                   "sign <filepath>",
		Short:                 "Sign the transaction in the file",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {

			infileStr := args[0]

			if err := cli.applyTxFile(infileStr); err != nil {
				fmt.Printf("Error apply infile(%s): %v\n", infileStr, err)
				return
			}

			fmt.Println("Transaction details are as follows:")
			cli.printTxIndent()

			var outStr string
			var err error
			if cmd.Flags().Changed("out") {
				outStr, err = cmd.Flags().GetString("out")
				if err != nil {
					fmt.Println(err)
				}
			}

			if outStr == "" {
				if infileStr == "" {
					outStr = "tx.sign"
				} else {
					outStr = infileStr + ".sign"
				}

			}
			cli.signTxAndSave(outStr)
		},
	}

	signTxCmd.Flags().String("out", "", "file `path` to save signed transaction")

	return signTxCmd
}

func (cli *CLI) buildBroadcastCmd() *cobra.Command {
	broadcastCmd := &cobra.Command{
		Use:                   "broadcast <signTxFilePath>",
		Short:                 "Broadcast sign transacion hex in the signTxFilePath to blockchain",
		Args:                  cobra.MinimumNArgs(1),
		Aliases:               []string{"submit"},
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			infileStr := args[0]

			signTxStr, err := readLineFromFile(infileStr)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(signTxStr))

			signTxByte := common.FromHex(string(signTxStr))
			signTx := new(types.Transaction)
			if err := rlp.DecodeBytes(signTxByte, signTx); err != nil {
				fmt.Println("DecodeBytes signTxHex error: ", err)
				return
			}

			ctx := context.Background()
			client, err := rpc.DialContext(ctx, cli.rpcURL)
			if err != nil {
				fmt.Println("DialContext: ", err)
				return
			}
			if err := client.CallContext(ctx, nil, "eth_sendRawTransaction", signTxStr); err != nil {
				fmt.Println("CallContext Error: ", err)
				return
			}
			fmt.Println("Waiting for transaction receipt...")
			waitMined(ctx, client, signTx.Hash())
			showTransactionReceipt(cli.rpcURL, signTx.Hash().String())
		},
	}
	return broadcastCmd
}

func (cli *CLI) buildSignMesgCmd() *cobra.Command {
	signMesgCmd := &cobra.Command{
		Use:                   "mesg [message] [--in infilepath] [--out outfilepath]",
		Short:                 "sign message or sign file",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	return signMesgCmd
}

func waitMined(ctx context.Context, client *rpc.Client, hash common.Hash) {
	transactionReceipt := func() (*types.Receipt, error) {
		var r *types.Receipt
		err := client.CallContext(ctx, &r, "eth_getTransactionReceipt", hash)
		if err == nil {
			if r == nil {
				return nil, ethereum.NotFound
			}
		}
		return r, err
	}

	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()

	for {
		receipt, err := transactionReceipt()
		if receipt != nil {
			break
		}
		if err != nil {
			// logger.Trace("Receipt retrieval failed", "err", err)
		} else {
			// logger.Trace("Transaction not yet mined")
		}
		// Wait for the next round.
		select {
		case <-ctx.Done():
			break
		case <-queryTicker.C:
		}
	}
}

func (cli *CLI) saveTranToFile(filepath string) error {
	tByte, err := cli.tran.MarshalJSON()
	if err != nil {
		return err
	}

	return saveByteToFile(tByte, filepath)
}

func saveByteToFile(b []byte, filepath string) error {
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	bN := append(b, '\n')
	if _, err := f.Write(bN); err != nil {
		return err
	}

	return nil
}

func saveStringToFile(str, filepath string) error {
	return saveByteToFile([]byte(str), filepath)
}

func readLineFromFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		text := scanner.Text()
		if len(text) > 0 {
			return text, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}

func (cli *CLI) signTxAndSave(filepath string) {
	signTx, err := cli.unlockAndSignTx()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Signed Transaction Hash: ", signTx.Hash().String())

	data, err := rlp.EncodeToBytes(signTx)
	if err != nil {
		fmt.Println(err)
		return
	}
	dataHex := common.Bytes2Hex(data) // common.ToHex(data)
	fmt.Printf("Signed Transaction: %s\n", dataHex)

	if err := saveStringToFile(dataHex, filepath); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Successfully save signed transacion hex to file", filepath)
}
