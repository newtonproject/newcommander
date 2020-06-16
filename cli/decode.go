package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spf13/cobra"
)

type TxData struct {
	From  common.Address  `json:"from"`
	To    *common.Address `json:"to"`
	Value string          `json:"value"`
	Data  string          `json:"data"`

	Nonce    uint64   `json:"nonce"    gencodec:"required"`
	Price    *big.Int `json:"gasPrice" gencodec:"required"`
	GasLimit uint64   `json:"gas"      gencodec:"required"`

	// Signature values
	V string `json:"v" gencodec:"required"`
	R string `json:"r" gencodec:"required"`
	S string `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash common.Hash `json:"hash" rlp:"-"`

	ChainID *big.Int `json:"chainID"`
}

func (cli *CLI) buildDecodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "decode <hexRawTransaction>",
		Short:                 "Decode hex raw transaction to json",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {

			b := common.FromHex(args[0])
			if len(b) == 0 {
				fmt.Println("convert hex string to bytes error")
				return
			}

			tx := new(types.Transaction)
			err := rlp.DecodeBytes(b, tx)
			if err != nil {
				fmt.Println("Decode Bytes Error: ", err)
				return
			}

			var signer types.Signer = types.FrontierSigner{}
			if tx.Protected() {
				signer = types.NewEIP155Signer(tx.ChainId())
			}
			from, err := types.Sender(signer, tx)
			if err != nil {
				fmt.Println("Sender Error: ", err)
			}

			// txData := new(TxData)
			v, r, s := tx.RawSignatureValues()
			txData := &TxData{
				From:     from,
				To:       tx.To(),
				Value:    getWeiAmountTextByUnit(tx.Value(), UnitETH),
				Data:     hex.EncodeToString(tx.Data()),
				Nonce:    tx.Nonce(),
				Price:    tx.GasPrice(),
				GasLimit: tx.Gas(),
				V:        common.ToHex(v.Bytes()),
				R:        common.ToHex(r.Bytes()),
				S:        common.ToHex(s.Bytes()),
				Hash:     tx.Hash(),
				ChainID:  tx.ChainId(),
			}

			var out []byte
			compress, _ := cmd.Flags().GetBool("compress")
			if compress {
				out, err = json.Marshal(txData)
			} else {
				out, err = json.MarshalIndent(txData, "", " ")
			}

			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(out))

			return
		},
	}

	cmd.Flags().BoolP("compress", "C", false, "Compress the out json")

	return cmd
}
