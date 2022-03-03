package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	prompt2 "github.com/ethereum/go-ethereum/console/prompt"
	"github.com/spf13/cobra"
)

// Transaction for send Transaction
type Transaction struct {
	From        common.Address `json:"from"`
	To          common.Address `json:"to"`
	Value       *big.Int       `json:"value"`
	Unit        string         `json:"unit"`
	Data        []byte         `json:"data"`
	Nonce       uint64         `json:"nonce"`
	GasPrice    *big.Int       `json:"gasPrice"`
	GasLimit    uint64         `json:"gas"`
	NetworkID   *big.Int       `json:"networkID"`
	Password    string         `json:"password,omitempty"`
	GasPriceTip *big.Int       `json:"gasTips,omitempty"`
}

func (t Transaction) MarshalJSON() ([]byte, error) {
	type transaction struct {
		From      common.Address  `json:"from"`
		To        *common.Address `json:"to"`
		Value     string          `json:"value"`
		Unit      string          `json:"unit"`
		Data      *hexutil.Bytes  `json:"data"`
		Nonce     uint64          `json:"nonce"`
		GasPrice  *big.Int        `json:"gasPrice"`
		GasLimit  uint64          `json:"gas"`
		NetworkID *big.Int        `json:"networkID"`
		// Password  string          `json:"password,omitempty"`
	}
	var tran transaction
	tran.From = t.From
	tran.To = &t.To
	tran.Value = getWeiAmountTextByUnit(t.Value, t.Unit)
	tran.Unit = t.Unit
	tran.Data = (*hexutil.Bytes)(&t.Data)
	tran.Nonce = t.Nonce
	tran.GasPrice = t.GasPrice
	tran.GasLimit = t.GasLimit
	tran.NetworkID = t.NetworkID
	// tran.Password = t.Password

	return json.MarshalIndent(tran, "", " ")
}

func (t *Transaction) UnmarshalJSON(input []byte) error {
	type transaction struct {
		From      common.Address  `json:"from"`
		To        *common.Address `json:"to"`
		Value     string          `json:"value"`
		Unit      string          `json:"unit"`
		Data      *hexutil.Bytes  `json:"data"`
		Nonce     uint64          `json:"nonce"`
		GasPrice  *big.Int        `json:"gasPrice"`
		GasLimit  uint64          `json:"gas"`
		NetworkID *big.Int        `json:"networkID"`
		Password  string          `json:"password,omitempty"`
	}
	var tran transaction
	if err := json.Unmarshal(input, &tran); err != nil {
		return err
	}

	t.From = tran.From
	if tran.To != nil {
		t.To = *tran.To
	}
	t.Unit = tran.Unit
	amountWei, err := getAmountWei(tran.Value, tran.Unit)
	if err != nil {
		return err
	}
	t.Value = amountWei
	if tran.Data != nil {
		t.Data = *tran.Data
	}
	t.Nonce = tran.Nonce
	if tran.GasPrice != nil {
		t.GasPrice = tran.GasPrice
	}
	if tran.GasLimit >= 21000 {
		t.GasLimit = tran.GasLimit
	}
	if tran.NetworkID != nil {
		t.NetworkID = tran.NetworkID
	}
	if tran.Password != "" {
		t.Password = tran.Password
	}

	return nil
}

func (cli *CLI) applyTxCobra(cmd *cobra.Command, args []string) error {
	if cli.tran == nil {
		return errCliTranNil
	}

	if cmd.Flags().Changed("unit") {
		unitStr, err := cmd.Flags().GetString("unit")
		if err != nil {
			return err
		}
		if !stringInSlice(unitStr, UnitList) {
			return errIllegalUnit
		}
		cli.tran.Unit = unitStr
	}

	if len(args) <= 0 {
		if cli.tran.Value == nil {
			return errors.New("Error: requires at least 1 arg(s), only received 0")
		}
	} else if strings.ToLower(args[0]) != "all" {
		amountStr := args[0]
		amountWei, err := getAmountWei(amountStr, cli.tran.Unit)
		if err != nil {
			return errIllegalAmount
		}
		cli.tran.Value = amountWei
	}

	if cmd.Flags().Changed("from") {
		fromStr, err := cmd.Flags().GetString("from")
		if err != nil {
			return err
		}
		if !common.IsHexAddress(fromStr) {
			return errFromAddressIllegal
		}
		cli.tran.From = common.HexToAddress(fromStr)
	} else if (cli.tran.From == common.Address{}) {
		return errRequiredFromAddress
	}

	if cmd.Flags().Changed("to") {
		toStr, err := cmd.Flags().GetString("to")
		if err != nil {
			return err
		}
		if !common.IsHexAddress(toStr) {
			return errToAddressIllegal
		}
		cli.tran.To = common.HexToAddress(toStr)
	} else if (cli.tran.To == common.Address{}) {
		return errRequiredToAddress
	}

	if cmd.Flags().Changed("data") {
		data, err := cmd.Flags().GetString("data")
		if err != nil {
			return err
		}
		cli.tran.Data = []byte(data)
	}

	return nil
}

func (cli *CLI) applyTxFile(path string) error {
	if cli.tran == nil {
		return errCliTranNil
	}
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return cli.tran.UnmarshalJSON(f)
}

func (cli *CLI) applyTxGuide(offline bool) error {
	var prompt string

	if cli.tran == nil {
		return errCliTranNil
	}

	// get from address
	for i := 0; ; i++ {
		if err := func() error {
			if cli.tran.From == (common.Address{}) {
				prompt = fmt.Sprintf("Enter from address who sign tx: ")
			} else {
				prompt = fmt.Sprintf("Enter from address who sign tx (default: %s): ", cli.tran.From.String())
			}
			fromAddressStr, err := prompt2.Stdin.PromptInput(prompt)
			if err != nil {
				return fmt.Errorf("Error: get \"from\" error")
			}
			if fromAddressStr == "" {
				if cli.tran.From == (common.Address{}) {
					return errRequiredFromAddress
				}
			} else {
				if !common.IsHexAddress(fromAddressStr) {
					return errFromAddressIllegal
				}
				cli.tran.From = common.HexToAddress(fromAddressStr)
			}
			return nil
		}(); err == nil {
			break
		} else if i < 2 {
			fmt.Println(err)
			continue
		} else {
			return err
		}
	}

	// get to address
	for i := 0; ; i++ {
		if err := func() error {
			if cli.tran.To == (common.Address{}) {
				prompt = fmt.Sprintf("Enter to address: ")
			} else {
				prompt = fmt.Sprintf("Enter to address (default: %s): ", cli.tran.To.String())
			}
			toAddressStr, err := prompt2.Stdin.PromptInput(prompt)
			if err != nil {
				return fmt.Errorf("Error: get \"to\" error")
			}
			if toAddressStr == "" {
				if cli.tran.To == (common.Address{}) {
					return errRequiredToAddress
				}
			} else {
				if !common.IsHexAddress(toAddressStr) {
					return errToAddressIllegal
				}
				cli.tran.To = common.HexToAddress(toAddressStr)
			}
			return nil
		}(); err == nil {
			break
		} else if i < 2 {
			fmt.Println(err)
			continue
		} else {
			return err
		}
	}

	// get pay amount unit
	for i := 0; ; i++ {
		if err := func() error {
			if cli.tran.Unit == "" {
				prompt = fmt.Sprintf("Enter unit for amount (%s or %s, default %s): ", UnitETH, UnitWEI, UnitETH)
				cli.tran.Unit = UnitETH
			} else {
				prompt = fmt.Sprintf("Enter unit for amount (%s or %s, default %s): ", UnitETH, UnitWEI, cli.tran.Unit)
			}
			unit, err := prompt2.Stdin.PromptInput(prompt)
			if err != nil {
				fmt.Println("Error: get \"unit\" error")
				return err
			}
			if unit == "" {
				unit = cli.tran.Unit
			} else {
				if !stringInSlice(unit, UnitList) {
					return errIllegalUnit
				}
				cli.tran.Unit = unit
			}
			return nil
		}(); err == nil {
			break
		} else if i < 2 {
			fmt.Println(err)
			continue
		} else {
			return err
		}
	}

	// get pay amount
	for i := 0; ; i++ {
		if err := func() error {
			if cli.tran.Value == nil {
				cli.tran.Value = new(big.Int)
			}
			prompt = fmt.Sprintf("Enter amount to pay in %s (default: %s): ", cli.tran.Unit,
				getWeiAmountTextByUnit(cli.tran.Value, cli.tran.Unit))
			amountStr, err := prompt2.Stdin.PromptInput(prompt)
			if err != nil {
				fmt.Println("PromptInput err:", err)
				return err
			}
			if amountStr == "" {
				if cli.tran.Value == nil {
					return errAmount0
				}
			} else {
				if !IsDecimalString(amountStr) {
					return errIllegalAmount
				}
				value, err := getAmountWei(amountStr, cli.tran.Unit)
				if err != nil {
					return err
				}
				cli.tran.Value = value
			}
			return nil
		}(); err == nil {
			break
		} else if i < 2 {
			fmt.Println(err)
			continue
		} else {
			return err
		}
	}

	// get pay message
	for i := 0; ; i++ {
		if err := func() error {
			if len(cli.tran.Data) > 0 {
				fmt.Println("Current text message is: ", string(cli.tran.Data))
				prompt = fmt.Sprintf("Enter text message (default no change): ")
			} else {
				prompt = fmt.Sprintf("Enter text message (default is empty): ")
			}
			dataStr, err := prompt2.Stdin.PromptInput(prompt)
			if err != nil {
				return err
			}
			if dataStr != "" {
				cli.tran.Data = []byte(dataStr)
			}
			return nil
		}(); err == nil {
			break
		} else if i < 2 {
			fmt.Println(err)
			continue
		} else {
			return err
		}
	}

	if !offline {
		return nil
	}

	// get nonce
	prompt = fmt.Sprintf("Enter nonce of from address (default: %d): ", cli.tran.Nonce)
	nonceStr, err := prompt2.Stdin.PromptInput(prompt)
	if err != nil {
		return err
	}
	if nonceStr != "" {
		nonce, err := strconv.ParseUint(nonceStr, 10, 64)
		if err != nil {
			return err
		}
		cli.tran.Nonce = nonce
	}

	// get gasPrice
	if cli.tran.GasPrice == nil {
		cli.tran.GasPrice = big.NewInt(1)
	}
	prompt = fmt.Sprintf("Enter gasPrice (default: %s WEI): ", cli.tran.GasPrice.String())
	gasPriceStr, err := prompt2.Stdin.PromptInput(prompt)
	if err != nil {
		fmt.Println("get gasPrice error")
		return err
	}
	if gasPriceStr != "" {
		if !IsDecimalString(gasPriceStr) {
			return errors.New("gasPrice invaild")
		}
		gasPrice, ok := new(big.Int).SetString(gasPriceStr, 10)
		if !ok {
			return errors.New("conver gasPrice to bigInt error")
		}
		cli.tran.GasPrice = gasPrice
	}

	// get GasLimit
	if cli.tran.GasLimit < 21000 {
		cli.tran.GasLimit = 21000
	}
	prompt = fmt.Sprintf("Enter gasLimit (default: %d): ", cli.tran.GasLimit)
	gasLimitStr, err := prompt2.Stdin.PromptInput(prompt)
	if err != nil {
		return fmt.Errorf("get gasLimit error")
	}
	if gasLimitStr != "" {
		gasLimit, err := strconv.ParseUint(gasLimitStr, 10, 64)
		if err != nil {
			return fmt.Errorf("conver gasLimit error")
		}
		cli.tran.GasLimit = gasLimit
	}

	// get ChainID
	if cli.tran.NetworkID == nil || cli.tran.NetworkID.Cmp(big.NewInt(0)) == 0 {
		cli.tran.NetworkID = DefaultChainID
	}
	prompt = fmt.Sprintf("Enter ChainID (default: %s): ", cli.tran.NetworkID.String())
	networkIDStr, err := prompt2.Stdin.PromptInput(prompt)
	if err != nil {
		return err
	}
	if networkIDStr != "" {
		networkID, ok := new(big.Int).SetString(networkIDStr, 10)
		if !ok {
			return errors.New("chainID conver error")
		}
		cli.tran.NetworkID = networkID
	}

	return nil
}
