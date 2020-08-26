package tracer

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

//go:generate go-bindata -pkg tracer -o gen_tracer.go tracer.js
//

func TracerJS() ([]byte, error) {
	return tracerJsBytes()
}

/*
[
    {
        "type":"call",
        "callType":"call",
        "from":"0x97549e368acafdcae786bb93d98379f1d1561a29",
        "to":"0x570611ba2d46ff0aca9f96168c4acbdd27bb0c54",
        "input":"0x2e1a7d4d0000000000000000000000000000000000000000000000000000000000000400",
        "output":"0x",
        "traceAddress":[],
        "value":"0x0",
        "gas":"0x3d59",
        "gasUsed":"0x3a69"
    }
]
*/

//go:generate gencodec -type Tx -field-override txMarshaling -out gen_tx_json.go

type Tx struct {
	From   common.Address  `db:"from"`
	To     *common.Address `db:"to"`
	Input  []byte          `db:"input"`
	Output []byte          `db:"output"`
	Value  *big.Int        `db:"value"`
}

type txMarshaling struct {
	From   common.Address
	To     *common.Address
	Input  hexutil.Bytes
	Output hexutil.Bytes
	Value  *hexutil.Big
}

type TraceConfig struct {
	Tracer  string  `json:"tracer,omitempty"`
	Timeout *string `json:"timeout,omitempty"`
	Reexec  *uint64 `json:"reexec,omitempty"`
}

func TraceTransaction(c *rpc.Client, ctx context.Context, txHash common.Hash, config *TraceConfig) ([]*Tx, error) {
	tjs, err := TracerJS()
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = &TraceConfig{
			Tracer: string(tjs),
		}
	}
	if config.Tracer == "" {
		config.Tracer = string(tjs)
	}

	var raw json.RawMessage
	err = c.CallContext(ctx, &raw, "debug_traceTransaction", txHash, config)
	if err != nil {
		return nil, err
	} else if len(raw) == 0 {
		return nil, ethereum.NotFound
	}

	var txs []*Tx

	err = json.Unmarshal(raw, &txs)
	if err != nil {
		return nil, err
	}

	return txs, nil
}
