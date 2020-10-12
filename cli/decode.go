package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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

	Raw               string `json:"raw" rlp:"-"`
	UnsignedRawTx     string `json:"UnsignedRawTx" rlp:"-"`
	UnsignedRawTxHash string `json:"UnsignedRawTxHash" rlp:"-"`
	PublicKey         string `json:"PublicKey" rlp:"-"`
}

func (cli *CLI) buildDecodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "decode <hexRawTransaction> [--rlp]",
		Short:                 "Decode hex raw transaction to json",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {

			b := common.FromHex(args[0])
			if len(b) == 0 {
				fmt.Println("convert hex string to bytes error")
				return
			}

			onlyRlp, _ := cmd.Flags().GetBool("rlp")
			compress, _ := cmd.Flags().GetBool("compress")

			tx := new(types.Transaction)
			err := rlp.DecodeBytes(b, tx)
			if err != nil {
				fmt.Println("Decode Bytes Error: ", err)
				return
			}

			if onlyRlp {
				var out []byte
				if compress {
					out, err = json.Marshal(tx)
				} else {
					out, err = json.MarshalIndent(tx, "", " ")
				}
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(string(out))

				return
			}

			var signer types.Signer = types.FrontierSigner{}
			if tx.Protected() {
				signer = types.NewEIP155Signer(tx.ChainId())
			}
			from, err := types.Sender(signer, tx)
			if err != nil {
				fmt.Println("Sender Error: ", err)
				return
			}

			tx0b, err := rlp.EncodeToBytes([]interface{}{
				tx.Nonce(),
				tx.GasPrice(),
				tx.Gas(),
				tx.To(),
				tx.Value(),
				tx.Data(),
				tx.ChainId(), uint(0), uint(0),
			})
			if err != nil {
				fmt.Println("tx0 rlp EncodeToBytes error: ", err)
				return
			}
			tx0 := new(types.Transaction)
			err = rlp.DecodeBytes(tx0b, tx0)
			if err != nil {
				fmt.Println("tx0 rlp DecodeBytes error: ", err)
				return
			}
			tx0h := signer.Hash(tx0)

			// txData := new(TxData)
			v, r, s := tx.RawSignatureValues()

			vb := big.NewInt(0).Set(v)
			if tx.Protected() {
				chainIdMul := new(big.Int).Mul(tx.ChainId(), big.NewInt(2))
				vb.Sub(v, chainIdMul)
				vb.Sub(vb, big.NewInt(8))
			}
			pubkey, err := recoverPublicKey(tx0.Hash(), r, s, vb, true)
			if err != nil {
				fmt.Println("recoverPublicKey error: ", err)
				return
			}
			pk, err := crypto.UnmarshalPubkey(pubkey)
			if err != nil {
				fmt.Println(err)
				return
			}
			if from != crypto.PubkeyToAddress(*pk) {
				fmt.Println("recover public key error")
				return
			}

			txData := &TxData{
				From:              from,
				To:                tx.To(),
				Value:             getWeiAmountTextByUnit(tx.Value(), UnitETH),
				Data:              hex.EncodeToString(tx.Data()),
				Nonce:             tx.Nonce(),
				Price:             tx.GasPrice(),
				GasLimit:          tx.Gas(),
				V:                 hexutil.Encode(v.Bytes()),
				R:                 hexutil.Encode(r.Bytes()),
				S:                 hexutil.Encode(s.Bytes()),
				Hash:              tx.Hash(),
				ChainID:           tx.ChainId(),
				Raw:               hexutil.Encode(b),
				UnsignedRawTx:     hexutil.Encode(tx0b),
				UnsignedRawTxHash: tx0h.String(),
				PublicKey:         hexutil.Encode(pubkey),
			}

			var out []byte
			if compress {
				out, err = json.Marshal(txData)
			} else {
				out, err = json.MarshalIndent(txData, "", " ")
			}
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("The raw transaction is decoded as follow:")
			fmt.Println(string(out))

			if compress {
				out, err = json.Marshal(tx0)
			} else {
				out, err = json.MarshalIndent(tx0, "", " ")
			}
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("The unsigned tx is decoded as follow:")
			fmt.Println(string(out))

			return
		},
	}

	cmd.Flags().BoolP("compress", "C", false, "Compress the out json")
	cmd.Flags().Bool("rlp", false, "Only decode rlp")

	return cmd
}

func recoverPublicKey(sighash common.Hash, R, S, Vb *big.Int, homestead bool) ([]byte, error) {
	if Vb.BitLen() > 8 {
		return nil, errors.New("invalid transaction v, r, s values")
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
		return nil, errors.New("invalid transaction v, r, s values")
	}
	// encode the signature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the signature
	pub, err := crypto.Ecrecover(sighash[:], sig)
	if err != nil {
		return nil, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return nil, errors.New("invalid public key")
	}

	return pub, nil
}
