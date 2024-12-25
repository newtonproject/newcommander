package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/sha3"
)

func (cli *CLI) buildBlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "block [blockNumber|latest|list]",
		Short:                 "Get the block signer of NewChain",
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Run: func(cmd *cobra.Command, args []string) {
			var number *big.Int
			if len(args) > 0 && args[0] != "latest" {
				var ok bool
				number, ok = big.NewInt(0).SetString(args[0], 10)
				if !ok {
					fmt.Println("Number error")
					return
				}
			}

			client, err := ethclient.Dial(cli.rpcURL)
			if err != nil {
				fmt.Println(err)
				return
			}
			ctx := context.Background()
			block, err := client.BlockByNumber(ctx, number)
			if err != nil {
				fmt.Println(err)
				return
			}

			address, err := ecrecover(block.Header())
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(block.Number().String(), address.String())

			return
		},
	}

	cmd.AddCommand(cli.buildBlockListCmd())

	return cmd
}

func (cli *CLI) buildBlockListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "list <number>",
		Short:                 "List the signers of the last <number> blocks",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			var number *big.Int
			if len(args) > 0 && args[0] != "latest" {
				var ok bool
				number, ok = big.NewInt(0).SetString(args[0], 10)
				if !ok {
					fmt.Println("Number error")
					return
				}
			}
			if number == nil || number.Cmp(big.NewInt(0)) <= 0 {
				number = big.NewInt(1)
			}

			client, err := ethclient.Dial(cli.rpcURL)
			if err != nil {
				fmt.Println(err)
				return
			}
			ctx := context.Background()
			latest, err := client.BlockByNumber(ctx, nil)
			if err != nil {
				fmt.Println(err)
				return
			}
			address, err := ecrecover(latest.Header())
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(latest.Number().String(), address.String())

			showClique, _ := cmd.Flags().GetBool("clique")
			var (
				nonceAuthVote = hexutil.MustDecode("0xffffffffffffffff") // Magic nonce number to vote on adding a new signer
				nonceDropVote = hexutil.MustDecode("0x0000000000000000") // Magic nonce number to vote on removing a signer.
			)

			for n := big.NewInt(1); n.Cmp(number) < 0; n.Add(n, big.NewInt(1)) {
				block, err := client.BlockByNumber(ctx, big.NewInt(0).Sub(latest.Number(), n))
				if err != nil {
					fmt.Println(err)
					return
				}
				address, err := ecrecover(block.Header())
				if err != nil {
					fmt.Println(err)
					return
				}
				if showClique {
					if block.Coinbase() != (common.Address{}) {
						if bytes.Compare(nonceAuthVote, block.Header().Nonce[:]) == 0 {
							fmt.Println(block.Number().String(), address.String(), "Auth", block.Coinbase().String())
						} else if bytes.Compare(nonceDropVote, block.Header().Nonce[:]) == 0 {
							fmt.Println(block.Number().String(), address.String(), "Drop", block.Coinbase().String())
						}
					} else {
						fmt.Println(block.Number().String(), address.String())
					}

				} else {
					fmt.Println(block.Number().String(), address.String())
				}
			}

			return
		},
	}

	cmd.Flags().Bool("clique", false, "show clique proposals")

	return cmd
}

// sealHash returns the hash of a block prior to it being sealed.
func sealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	encodeSigHeader(hasher, header)
	hasher.(crypto.KeccakState).Read(hash[:])
	return hash
}

func encodeSigHeader(w io.Writer, header *types.Header) {
	enc := []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra[:len(header.Extra)-crypto.SignatureLength], // Yes, this will panic if extra is too short
		header.MixDigest,
		header.Nonce,
	}
	if header.BaseFee != nil {
		enc = append(enc, header.BaseFee)
	}
	// if header.WithdrawalsHash != nil {
	// 	panic("unexpected withdrawal hash value in clique")
	// }
	// if header.ExcessBlobGas != nil {
	// 	panic("unexpected excess blob gas value in clique")
	// }
	// if header.BlobGasUsed != nil {
	// 	panic("unexpected blob gas used value in clique")
	// }
	// if header.ParentBeaconRoot != nil {
	// 	panic("unexpected parent beacon root value in clique")
	// }
	if err := rlp.Encode(w, enc); err != nil {
		panic("can't encode: " + err.Error())
	}
}

// ecrecover extracts the Ethereum account address from a signed header.
func ecrecover(header *types.Header) (common.Address, error) {
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < 65 {
		return common.Address{}, errors.New("extra-data 65 byte signature suffix missing")
	}
	signature := header.Extra[len(header.Extra)-65:]

	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(sealHash(header).Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

	return signer, nil
}
