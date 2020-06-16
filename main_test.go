package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/newtonproject/newcommander/cli"
	"github.com/sirupsen/logrus"
)

// test address
const taddr = "0xDB2C9C06E186D58EFe19f213b3d5FaF8B8c99481"

func getTempFile() (string, func()) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	file := dir + string(os.PathSeparator) + "lumen-integration-test.json"

	return file, func() {
		logrus.Debugf("cleaning up temp file: %s", file)
		os.RemoveAll(dir)
	}
}

func run(cli *cli.CLI, command string) string {
	fmt.Printf("$ ./newcommander %s\n", command)
	got := cli.TestCommand(command)
	fmt.Printf("%s\n", got)
	return strings.TrimSpace(got)
}

func runArgs(cli *cli.CLI, args ...string) string {
	fmt.Printf("$ ./newcommander %s\n", strings.Join(args, " "))
	got := cli.Embeddable().Run(args...)
	fmt.Printf("%s\n", got)
	return strings.TrimSpace(got)
}

func expectOutput(t *testing.T, cli *cli.CLI, want string, command string) {
	got := run(cli, command)

	if got != want {
		t.Errorf("(%s) wrong output: want %v, got %v", command, want, got)
	}
}

func newCLI() (*cli.CLI, func()) {
	_, cleanupFunc := getTempFile()

	glumen := cli.NewCLI()
	glumen.TestCommand("version")
	run(glumen, fmt.Sprintf("version"))

	return glumen, cleanupFunc
}

func getBalance(cli *cli.CLI, account string) float64 {
	balanceString := run(cli, "balance "+account)

	balance, err := strconv.ParseFloat(balanceString, 64)

	if err != nil {
		return 0
	}

	return balance
}

// Create new funded test account
func TestTx(t *testing.T) {
	cli, _ := newCLI()
	address := run(cli, "account new") // no password

	balance := getBalance(cli, address)
	if balance <= 100 {
		run(cli, fmt.Sprintf("pay 1 --from %s --to %s", taddr, address))
		// wait for tx??
	}

	balance = getBalance(cli, address)
	if balance < 0 {
		t.Fatalf("expected balance < 100 got %v", balance)
	}
}
