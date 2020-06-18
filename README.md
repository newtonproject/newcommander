## newcommander

This is a commandline client for the NewChain blockchain.
It's designed to be easy-to-use. It contains the following:
* Create a new account
* Get the balance
* Send transactions to the NewChain system
* Off-line sign transaction from file or guide
* Online send signed transaction
* RPC method support
* Batch pay

## QuickStart

### Download from releases

Binary archives are published at:
* [NewChain](https://release.cloud.diynova.com/newton/NewCommander/)
* [Ethereum](https://release.cloud.diynova.com/newton/NewCommander/ethereum/)

### Building the source

To get from gitlab via `go get`, this will get source and install dependens(cobra, viper, logrus).

#### Windows

install command

```bash
git clone https://github.com/newtonproject/newcommander.git && cd newcommander && make install
```

run newcommander

```bash
%GOPATH%/bin/newcommander.exe
```

#### Linux or Mac

install:

```bash
git clone https://github.com/newtonproject/newcommander.git && cd newcommander && make install
```
run newcommander

```bash
$GOPATH/bin/newcommander
```

### Usage

#### Help

Use command `newcommander help` to display the usage.

```bash
Usage:
  newcommander [flags]
  newcommander [command]

Available Commands:
  account     Manage NewChain accounts
  balance     Get balance of address
  batchpay    Batch pay base on file <batch.txt>
  broadcast   Broadcast sign transacion hex in the signTxFilePath to blockchain
  build       Build transaction
  decode      Decode hex raw transaction to json
  faucet      Get free money for address on NewChain TestNet
  help        Help about any command
  init        Initialize config file
  pay         Send [amount] [unit] from [source] to [target] with message [text]
  rpc         NewChain RPC method
  sign        Sign the transaction in the file
  version     Get version of newcommander CLI

Flags:
  -c, --config path            The path to config file (default "./config.toml")
  -h, --help                   help for newcommander
  -i, --rpcURL url             NewChain json rpc or ipc url (default "https://rpc1.newchain.newtonproject.org")
  -w, --walletPath directory   Wallet storage directory (default "./wallet/")

Use "newcommander [command] --help" for more information about a command.
```

#### Use config.toml

You can use a configuration file to simplify the command line parameters.

One available configuration file `config.toml` is as follows:


```conf
rpcurl = "https://rpc1.newchain.newtonproject.org"
walletpath = "./wallet/"

[pay]
  from = "0xDB2C9C06E186D58EFe19f213b3d5FaF8B8c99481"
  unit = "NEW"
  password = "newton"
```

#### Initialize config file

```bash
# Initialize config file
newcommander init
```

Just press Enter to use the default configuration, and it's best to create a new user.

```bash
$ newcommander init
Initialize config file
Enter file in which to save (./config.toml):
Enter the wallet storage directory (./wallet/):
Enter NewChain json rpc or ipc url (https://rpc1.newchain.newtonproject.org):
Create a new account or not: [Y/n]
Your new account is locked with a password. Please give a password. Do not forget this password.
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
0xDB2C9C06E186D58EFe19f213b3d5FaF8B8c99481
Your configuration has been saved in  ./config.toml
```

#### Create account

```bash
# Create an account
newcommander account new

# Create 10 accounts
newcommander account new -n 10

# Create an account with the standard scrypt for keystore
newcommander account new -s
```

### List all accounts

```bash
# list all accounts of the walletPath
newcommander account list
```

### Update an account

```bash
# update an account password
newcommander account update 0x0e0D78D2089F577d8b8156Eab564f08Ec2249b30

# update account with the standard scrypt for keystore
newcommander account update 0x0e0D78D2089F577d8b8156Eab564f08Ec2249b30 -s
```

### Get faucet

```bash
# Get free money for address
newcommander faucet 0xDB2C9C06E186D58EFe19f213b3d5FaF8B8c99481
```

### Get Balance

```bash
# Get the balance of an account
newcommander balance 0xDB2C9C06E186D58EFe19f213b3d5FaF8B8c99481

# Get the balance of multiple accounts
newcommander balance 0xDB2C9C06E186D58EFe19f213b3d5FaF8B8c99481 0x25a03e72bcca5ddcda45a7aabbe9b41b0e8ff828

# Get the balance of all accounts in the default walletPath directory
newcommander balance

# Get the balance in the specified unit
newcommander balance -u NEW

# Get the balance at safe mode (3 block less than the latest)
newcommander balance -s

# Get the balance at specified height
newcommander balance -n 1024

# Get the pending balance
newcommander balance -n pending
```

### Pay to an account
```bash
# Pay 1 NEW to an account
newcommander pay 1 --to 0x25a03e72bcca5ddcda45a7aabbe9b41b0e8ff828

# Pay 1 WEI to an account
newcommander pay 1 -u WEI --to 0x25a03e72bcca5ddcda45a7aabbe9b41b0e8ff828

# Pay 1 NEW to an account with custom gasPrice 1000 WEI
newcommander pay 1 --to 0x25a03e72bcca5ddcda45a7aabbe9b41b0e8ff828 -p 1000

# Pay 1 NEW to an account with custom gasLimit 100000
newcommander pay 1 --to 0x25a03e72bcca5ddcda45a7aabbe9b41b0e8ff828 -g 100000

# Pay 1 NEW to an account with custom nonce 1
newcommander pay 1 --to 0x25a03e72bcca5ddcda45a7aabbe9b41b0e8ff828 -n 1

# Pay entire balance to an account
newcommander pay all --to 0x25a03e72bcca5ddcda45a7aabbe9b41b0e8ff828

# Pay to an account with txs confusion
newcommander pay 1 --to 0x25a03e72bcca5ddcda45a7aabbe9b41b0e8ff828 -N 2 -X 20
```

### Build transaction
```bash
# Build transaction
newcommander build
```

### Sign transaction
```bash
# Sign transaction from file and save sign transaction hex to file
newcommander sign tx.txt

# Sign transaction save to the specified file
newcommander sign tx.txt --out tx.sign
```

### Submit signed transaction
```bash
# Submit signed transaction hex to NewChain system
newcommander submit tx.sign
```

### Decode singed transaction
```bash
# Decode signed transaction hex string to json
newcommander decode 0xf863820258648252089497549e368acafdcae786bb93d98379f1d1561a298080820802a0768ff39803904e993df858e0d4bbc2d56adf37e804e6804e9e2532d6726c70a4a02e9d8a4cdaf1a8d1b7d99f783162bcd8f0a80084b069f6c69c77f4345c4392f8

# Decode signed transaction hex string to compress json
newcommander decode 0xf863820258648252089497549e368acafdcae786bb93d98379f1d1561a298080820802a0768ff39803904e993df858e0d4bbc2d56adf37e804e6804e9e2532d6726c70a4a02e9d8a4cdaf1a8d1b7d99f783162bcd8f0a80084b069f6c69c77f4345c4392f8 --compress
```

### Batch pay
```bash
# Batch pay base on batch.txt
newcommander batch batch.txt
newcommander batchpay batch.txt
```

### RPC
```bash
# Get chainID/NewworkID
newcommander rpc net_version

# Get block by number
newcommander rpc eth_getBlockByNumber 1024 true

# Get latest block
newcommander rpc eth_getBlockByNumber latest true

# Get transaction by hash
newcommander rpc eth_getTransactionByHash 0xcad4299fd6516c7f66cbb5ae70114d9c06d73908a66c9165dbfc1e36fb67d892

# Get transaction receipt
newcommander rpc eth_getTransactionReceipt 0xd9f1a3a4c54b6218c848e9a246faa570ea592a9f28d8b14e3ad4035398875de7

# Get account nonce
newcommander rpc eth_getTransactionCount 0xc94770007dda54cF92009BFF0dE90c06F603a09f latest

# Get balance
newcommander rpc eth_getBalance 0xc94770007dda54cF92009BFF0dE90c06F603a09f latest

# Get gas limit
newcommander rpc eth_estimateGas {}

# Get tranfer gas limit
newcommander rpc eth_estimateGas '{"from":"0xdf9106238879143e914ad78d4ff4b4fa6b3b1648","gasPrice":"0x64","to":"0xdf9106238879143e914ad78d4ff4b4fa6b3b1648","value":"0xde0b6b3a7640000"}'

# Get contract gas limit
newcommander rpc eth_estimateGas '{"data":"0xa9059cbb00000000000000000000000082a3a88bc9d6a70c4f3c66534566892eae0cad810000000000000000000000000000000000000000000000000000000000000001","from":"0x82a3a88bc9d6a70c4f3c66534566892eae0cad89","to":"0x0b7789b5f69678f4f2d237cd0e1c815e1cd39ccf","value":"0x0"}'

# Get gas price
newcommander rpc eth_gasPrice

# Send raw transaction
newcommander rpc eth_sendRawTransaction 0xf869206482520894e78a7560107831726517f0a7c64600b01572967d880de0b6b3a764000080828414a0d8c63158d7c2f4acc556b06d0b3f90e36cf57ef3ab1915a70b5bcf9e56978b63a0435430e53af79a76de3bd50d0f740f28fe983aa2ac494feb6529dc08c0fdc4ed

# Send transactions
newcommander rpc eth_sendTransaction '{"from":"0xb60e8dd61c5d32be8058bb8eb970870f07233155","to":"0xd46e8dd67c5d32be8058bb8eb970870f07244567","gas":"0x76c0","gasPrice":"0x9184e72a000","value":"0x9184e72a","data":"0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"}'
```
