package cli

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
				fmt.Println(block.Number().String(), address.String())
			}

			return
		},
	}

	return cmd
}

func sigHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	rlp.Encode(hasher, []interface{}{
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
		header.Extra[:len(header.Extra)-65], // Yes, this will panic if extra is too short
		header.MixDigest,
		header.Nonce,
	})
	hasher.Sum(hash[:0])
	return hash
}

// ecrecover extracts the Ethereum account address from a signed header.
func ecrecover(header *types.Header) (common.Address, error) {
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < 65 {
		return common.Address{}, errors.New("extra-data 65 byte signature suffix missing")
	}
	signature := header.Extra[len(header.Extra)-65:]

	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(sigHash(header).Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

	return signer, nil
}
