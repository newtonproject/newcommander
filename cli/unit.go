package cli

import (
	"fmt"
	"strings"
)

type Unit string

var (
	UnitETH = "ETH"
	UnitWEI = "WEI"

	// UnitList is array for Unit string
	// UnitList = []string{"Wei", "Ada", "Babbage", "Shannon", "Szabo", "Finney", "Ether", "Einstein", "Douglas", "Gwei"}
	UnitList []string

	// UnitString is for Unit string
	// UnitString = "Available unit: Wei, Ada, Babbage, Shannon, Szabo, Finney, Ether, Einstein, Douglas, Gwei"
	UnitString string
)

func InitUnit(bc BlockChain) {
	if bc == NewChain {
		UnitETH = "NEW"
		UnitWEI = "ISAAC"
	}

	UnitList = []string{UnitETH, UnitWEI}
	UnitString = fmt.Sprintf("Available unit: %s", strings.Join(UnitList, ","))
}
