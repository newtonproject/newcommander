package cli

import (
	"errors"
)

var (
	errCliTranNil          = errors.New("cli tran error")
	errBigSetString        = errors.New("convert string to big error")
	errIllegalAmount       = errors.New("illegal Amount")
	errIllegalUnit         = errors.New("illegal Unit")
	errRequiredFromAddress = errors.New(`required flag(s) "from" not set`)
	errFromAddressNil      = errors.New("from address nil")
	errFromAddressIllegal  = errors.New("from address illegal")
	errToAddressNil        = errors.New("to address nil")
	errToAddressIllegal    = errors.New("to address illegal")
	errWalletPathEmpty     = errors.New("empty wallet, create account first")
	errAmount0             = errors.New("not set send amount or amount is 0")
	errRequiredToAddress   = errors.New("not set to address")
)
