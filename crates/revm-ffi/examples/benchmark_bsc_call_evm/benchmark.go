package main

import (
	"fmt"
	"math/big"
	"time"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/core/rawdb"
)

// Contract bytecode - using the exact same working bytecode from FFI benchmark
const (
	// ERC20 with mint function (no constructor, tokens minted via function call) - same as FFI benchmark
	ERC20_WITH_MINT_BYTECODE = "6080604052348015600e575f5ffd5b506105348061001c5f395ff3fe608060405234801561000f575f5ffd5b506004361061004a575f3560e01c806318160ddd1461004e57806340c10f191461006c57806370a0823114610088578063a9059cbb146100b8575b5f5ffd5b6100566100e8565b60405161006391906102b6565b60405180910390f35b61008660048036038101906100819190610357565b6100ee565b005b6100a2600480360381019061009d9190610395565b61015c565b6040516100af91906102b6565b60405180910390f35b6100d260048036038101906100cd9190610357565b610170565b6040516100df91906103da565b60405180910390f35b60015481565b8060015f8282546100ff9190610420565b92505081905550805f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8282546101519190610420565b925050819055505050565b5f602052805f5260405f205f915090505481565b5f815f5f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205410156101f0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101e7906104ad565b60405180910390fd5b815f5f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825461023b91906104cb565b92505081905550815f5f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825461028d9190610420565b925050819055506001905092915050565b5f819050919050565b6102b08161029e565b82525050565b5f6020820190506102c95f8301846102a7565b92915050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6102fc826102d3565b9050919050565b61030c816102f2565b8114610316575f5ffd5b50565b5f8135905061032781610303565b92915050565b6103368161029e565b8114610340575f5ffd5b50565b5f813590506103518161032d565b92915050565b5f5f6040838503121561036d5761036c6102cf565b5b5f61037a85828601610319565b925050602061038b85828601610343565b9150509250929050565b5f602082840312156103aa576103a96102cf565b5b5f6103b784828501610319565b91505092915050565b5f8115159050919050565b6103d4816103c0565b82525050565b5f6020820190506103ed5f8301846103cb565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61042a8261029e565b91506104358361029e565b925082820190508082111561044d5761044c6103f3565b5b92915050565b5f82825260208201905092915050565b7f496e73756666696369656e742062616c616e63650000000000000000000000005f82015250565b5f610497601483610453565b91506104a282610463565b602082019050919050565b5f6020820190508181035f8301526104c48161048b565b9050919050565b5f6104d58261029e565b91506104e08361029e565b92508282039050818111156104f8576104f76103f3565b5b9291505056fea2646970667358221220cb19b4849bfce8663cd0287ac9f324dbeeb9b43d26a167e5a17d8452191f599d64736f6c634300081c0033"

	// Transfer amount (1 token with 18 decimals) - same as FFI benchmark
	TRANSFER_AMOUNT = "1000000000000000000" // 1 * 10^18
	TOTAL_SUPPLY    = "1000000000000000000000000000" // 1 billion tokens
)

// Test addresses - same as FFI benchmark
var (
	DEPLOYER_ADDRESS = common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96c4b4d8b6")
	ALICE_ADDRESS    = common.HexToAddress("0x8ba1f109551bD432803012645aac136c12345678")
	BOB_ADDRESS      = common.HexToAddress("0x1234567890123456789012345678901234567890")
	CHARLIE_ADDRESS  = common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
)

type EVMEnvironment struct {
	StateDB *state.StateDB
	EVM     *vm.EVM
	ChainID *big.Int
}

func main() {
	fmt.Println("🚀 Native Go EVM ERC20 Transfer Benchmark")
	fmt.Println("==========================================")

	// Test different configurations - same as FFI benchmark
	configs := []struct {
		name         string
		transferCount int
		chainID      *big.Int
	}{
		{"100 transfers on BSC Testnet", 100, big.NewInt(97)},
		{"1000 transfers on BSC Testnet", 1000, big.NewInt(97)},
		{"1000 transfers on Ethereum Mainnet", 1000, big.NewInt(1)},
	}

	for i, config := range configs {
		fmt.Printf("\n📊 Benchmark %d: %s\n", i+1, config.name)
		runBenchmark(config.transferCount, config.chainID)
	}
}

func runBenchmark(transferCount int, chainID *big.Int) {
	startTime := time.Now()

	// 1. Create EVM environment
	fmt.Print("   🔧 Creating EVM environment... ")
	env, err := createEVMEnvironment(chainID)
	if err != nil {
		fmt.Printf("❌ Failed to create EVM environment: %v\n", err)
		return
	}
	fmt.Printf("✅ (Chain ID: %d)\n", chainID.Int64())

	// 2. Setup accounts with initial balances
	fmt.Print("   💰 Setting up accounts... ")
	accountSetupTime := time.Now()
	setupAccounts(env)
	fmt.Printf("✅ (%v)\n", time.Since(accountSetupTime))

	// 3. Deploy ERC20 token
	fmt.Print("   📄 Deploying ERC20 token... ")
	deployTime := time.Now()
	
	contractAddress, err := deployContract(env, DEPLOYER_ADDRESS, ERC20_WITH_MINT_BYTECODE)
	if err != nil {
		fmt.Printf("❌ Failed to deploy contract: %v\n", err)
		return
	}
	
	fmt.Printf("✅ %s (%v)\n", contractAddress.Hex(), time.Since(deployTime))

	// 4. Mint tokens to deployer (same as FFI benchmark)
	fmt.Print("   🪙 Minting tokens to deployer... ")
	mintTime := time.Now()
	
	if err := mintTokens(env, DEPLOYER_ADDRESS, contractAddress, TOTAL_SUPPLY); err != nil {
		fmt.Printf("❌ Failed to mint tokens: %v\n", err)
		return
	}
	
	// Verify minting
	deployerBalance, _ := getTokenBalance(env, DEPLOYER_ADDRESS, contractAddress)
	fmt.Printf("✅ New balance: %s (%v)\n", formatTokenAmount(deployerBalance), time.Since(mintTime))

	// 5. Transfer tokens to Alice
	fmt.Print("   🎯 Setting up token balances... ")
	tokenSetupTime := time.Now()
	
	if err := transferTokens(env, DEPLOYER_ADDRESS, ALICE_ADDRESS, contractAddress, TOTAL_SUPPLY); err != nil {
		fmt.Printf("❌ Failed to transfer tokens to Alice: %v\n", err)
		return
	}
	
	// Verify transfer
	aliceBalance, _ := getTokenBalance(env, ALICE_ADDRESS, contractAddress)
	fmt.Printf("✅ Alice balance: %s (%v)\n", formatTokenAmount(aliceBalance), time.Since(tokenSetupTime))

	// 6. Perform benchmark transfers
	fmt.Printf("   🚀 Performing %d transfers... ", transferCount)
	transferTime := time.Now()

	success := performTransfers(env, contractAddress, transferCount)
	transferDuration := time.Since(transferTime)

	if !success {
		fmt.Println("❌ Transfers failed")
		return
	}

	fmt.Printf("✅ (%v)\n", transferDuration)

	// 7. Verify final balances
	fmt.Print("   ✅ Verifying balances... ")
	verifyTime := time.Now()
	
	finalAliceBalance, _ := getTokenBalance(env, ALICE_ADDRESS, contractAddress)
	finalBobBalance, _ := getTokenBalance(env, BOB_ADDRESS, contractAddress)
	
	fmt.Printf("✅ Alice: %s, Bob: %s (%v)\n", 
		formatTokenAmount(finalAliceBalance), 
		formatTokenAmount(finalBobBalance), 
		time.Since(verifyTime))

	// 8. Print summary
	totalTime := time.Since(startTime)
	transfersPerSecond := float64(transferCount) / transferDuration.Seconds()
	avgPerTransfer := transferDuration.Nanoseconds() / int64(transferCount)

	fmt.Printf("   📈 Summary:\n")
	fmt.Printf("      • Total time: %v\n", totalTime)
	fmt.Printf("      • Transfer time: %v\n", transferDuration)
	fmt.Printf("      • Transfers/second: %.2f\n", transfersPerSecond)
	fmt.Printf("      • Average per transfer: %.3fµs\n", float64(avgPerTransfer)/1000.0)
}

func createEVMEnvironment(chainID *big.Int) (*EVMEnvironment, error) {
	// Create state database
	stateDB, err := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	if err != nil {
		return nil, err
	}

	// Configure chain based on chain ID - use Cancun hardfork for PUSH0 support
	var chainCfg *params.ChainConfig
	if chainID.Cmp(big.NewInt(1)) == 0 {
		// Ethereum Mainnet - use Cancun hardfork for PUSH0 support
		chainCfg = &params.ChainConfig{
			ChainID:                       big.NewInt(1),
			HomesteadBlock:                big.NewInt(0),
			DAOForkBlock:                  nil,
			DAOForkSupport:                false,
			EIP150Block:                   big.NewInt(0),
			EIP155Block:                   big.NewInt(0),
			EIP158Block:                   big.NewInt(0),
			ByzantiumBlock:                big.NewInt(0),
			ConstantinopleBlock:           big.NewInt(0),
			PetersburgBlock:               big.NewInt(0),
			IstanbulBlock:                 big.NewInt(0),
			MuirGlacierBlock:              big.NewInt(0),
			BerlinBlock:                   big.NewInt(0),
			LondonBlock:                   big.NewInt(0),
			ArrowGlacierBlock:             big.NewInt(0),
			GrayGlacierBlock:              big.NewInt(0),
			MergeNetsplitBlock:            big.NewInt(0),
			ShanghaiTime:                  new(uint64),
			CancunTime:                    new(uint64), // Enable Cancun for PUSH0
		}
	} else {
		// BSC Testnet/Mainnet - use Cancun hardfork for PUSH0 support
		chainCfg = &params.ChainConfig{
			ChainID:                       chainID,
			HomesteadBlock:                big.NewInt(0),
			DAOForkBlock:                  nil,
			DAOForkSupport:                false,
			EIP150Block:                   big.NewInt(0),
			EIP155Block:                   big.NewInt(0),
			EIP158Block:                   big.NewInt(0),
			ByzantiumBlock:                big.NewInt(0),
			ConstantinopleBlock:           big.NewInt(0),
			PetersburgBlock:               big.NewInt(0),
			IstanbulBlock:                 big.NewInt(0),
			MuirGlacierBlock:              big.NewInt(0),
			BerlinBlock:                   big.NewInt(0),
			LondonBlock:                   big.NewInt(0),
			ArrowGlacierBlock:             big.NewInt(0),
			GrayGlacierBlock:              big.NewInt(0),
			MergeNetsplitBlock:            big.NewInt(0),
			ShanghaiTime:                  new(uint64),
			CancunTime:                    new(uint64), // Enable Cancun for PUSH0
		}
	}

	// Create block context
	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     func(uint64) common.Hash { return common.Hash{} },
		Coinbase:    common.Address{},
		BlockNumber: big.NewInt(1),
		Time:        uint64(time.Now().Unix()),
		Difficulty:  big.NewInt(1),
		BaseFee:     big.NewInt(1),
		GasLimit:    30000000,
	}

	// Create transaction context
	txCtx := vm.TxContext{
		Origin:   DEPLOYER_ADDRESS,
		GasPrice: big.NewInt(1),
	}

	// Create EVM
	evm := vm.NewEVM(blockCtx, txCtx, stateDB, chainCfg, vm.Config{})

	return &EVMEnvironment{
		StateDB: stateDB,
		EVM:     evm,
		ChainID: chainID,
	}, nil
}

func setupAccounts(env *EVMEnvironment) {
	// Set up accounts with ETH balances
	accounts := []common.Address{DEPLOYER_ADDRESS, ALICE_ADDRESS, BOB_ADDRESS, CHARLIE_ADDRESS}
	balance := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)) // 100 ETH

	for _, addr := range accounts {
		env.StateDB.SetBalance(addr, balance)
		env.StateDB.SetNonce(addr, 0)
	}
}

func deployContract(env *EVMEnvironment, from common.Address, bytecode string) (common.Address, error) {
	// Convert hex bytecode to bytes
	code := common.FromHex(bytecode)
	
	// Get nonce for contract address calculation
	nonce := env.StateDB.GetNonce(from)
	
	// Increment nonce
	env.StateDB.SetNonce(from, nonce+1)
	
	// Create a proper transaction context for deployment
	txCtx := vm.TxContext{
		Origin:   from,
		GasPrice: big.NewInt(1),
	}
	env.EVM.TxContext = txCtx
	
	// Create contract with more gas and check for errors
	ret, contractAddr, _, err := env.EVM.Create(vm.AccountRef(from), code, 30000000, big.NewInt(0))
	if err != nil {
		return common.Address{}, fmt.Errorf("contract creation failed: %v", err)
	}
	
	// Check if the contract was actually deployed
	deployedCode := env.StateDB.GetCode(contractAddr)
	if len(deployedCode) == 0 {
		return common.Address{}, fmt.Errorf("contract deployment failed: no code at address %s (ret len: %d)", contractAddr.Hex(), len(ret))
	}
	
	return contractAddr, nil
}

func mintTokens(env *EVMEnvironment, to, contract common.Address, amount string) error {
	// Encode mint(address,uint256) call
	// Function selector: 0x40c10f19
	
	// Parse amount - handle both hex and decimal strings
	var amountBig *big.Int
	if strings.HasPrefix(amount, "0x") {
		var ok bool
		amountBig, ok = new(big.Int).SetString(amount[2:], 16)
		if !ok {
			return fmt.Errorf("invalid hex amount: %s", amount)
		}
	} else {
		var ok bool
		amountBig, ok = new(big.Int).SetString(amount, 10)
		if !ok {
			return fmt.Errorf("invalid decimal amount: %s", amount)
		}
	}

	// Prepare call data - same encoding as FFI version
	input := make([]byte, 4+32+32) // selector + address + amount
	copy(input[0:4], []byte{0x40, 0xc1, 0x0f, 0x19}) // mint selector
	
	// Encode address (32 bytes, left-padded with zeros)
	copy(input[4+12:4+32], to.Bytes())
	
	// Encode amount (32 bytes, big-endian)
	amountBytes := amountBig.Bytes()
	copy(input[4+32+32-len(amountBytes):4+32+32], amountBytes)

	// Call from deployer address (same as FFI version)
	ret, err := callContract(env, DEPLOYER_ADDRESS, contract, input)
	if err != nil {
		return fmt.Errorf("mint call failed: %v", err)
	}
	
	// Check if mint was successful (should return empty for successful mint)
	if len(ret) != 0 {
		return fmt.Errorf("mint call returned unexpected data: %x", ret)
	}
	
	return nil
}

func transferTokens(env *EVMEnvironment, from, to, contract common.Address, amount string) error {
	// Encode transfer(address,uint256) call
	// Function selector: 0xa9059cbb
	
	// Parse amount - handle both hex and decimal strings
	var amountBig *big.Int
	if strings.HasPrefix(amount, "0x") {
		var ok bool
		amountBig, ok = new(big.Int).SetString(amount[2:], 16)
		if !ok {
			return fmt.Errorf("invalid hex amount: %s", amount)
		}
	} else {
		var ok bool
		amountBig, ok = new(big.Int).SetString(amount, 10)
		if !ok {
			return fmt.Errorf("invalid decimal amount: %s", amount)
		}
	}

	// Prepare call data - same encoding as FFI version
	input := make([]byte, 4+32+32) // selector + address + amount
	copy(input[0:4], []byte{0xa9, 0x05, 0x9c, 0xbb}) // transfer selector
	
	// Encode address (32 bytes, left-padded with zeros)
	copy(input[4+12:4+32], to.Bytes())
	
	// Encode amount (32 bytes, big-endian)
	amountBytes := amountBig.Bytes()
	copy(input[4+32+32-len(amountBytes):4+32+32], amountBytes)

	ret, err := callContract(env, from, contract, input)
	if err != nil {
		return fmt.Errorf("transfer call failed: %v", err)
	}
	
	// Check if transfer returned true (same as FFI version)
	if len(ret) == 32 && ret[31] == 1 {
		return nil
	}
	
	// Debug output for failed transfers
	if len(ret) > 0 {
		return fmt.Errorf("transfer returned false (got %d bytes: %x)", len(ret), ret)
	} else {
		return fmt.Errorf("transfer returned no data")
	}
}

func callContract(env *EVMEnvironment, from, contract common.Address, input []byte) ([]byte, error) {
	// Create a new transaction context for this call
	txCtx := vm.TxContext{
		Origin:   from,
		GasPrice: big.NewInt(1),
	}
	
	// Update the EVM with the new transaction context
	env.EVM.TxContext = txCtx
	
	// Perform the call with sufficient gas
	ret, _, err := env.EVM.Call(vm.AccountRef(from), contract, input, 1000000, big.NewInt(0))
	
	return ret, err
}

func getTokenBalance(env *EVMEnvironment, account, contract common.Address) (*big.Int, error) {
	// Encode balanceOf(address) call
	// Function selector: 0x70a08231
	
	input := make([]byte, 4+32) // selector + address
	copy(input[0:4], []byte{0x70, 0xa0, 0x82, 0x31}) // balanceOf selector
	
	// Encode address (32 bytes, left-padded)
	copy(input[4+12:4+32], account.Bytes())

	ret, err := callContract(env, DEPLOYER_ADDRESS, contract, input)
	if err != nil {
		return big.NewInt(0), err
	}
	
	if len(ret) != 32 {
		return big.NewInt(0), fmt.Errorf("invalid balance response")
	}
	
	return new(big.Int).SetBytes(ret), nil
}

func performTransfers(env *EVMEnvironment, contract common.Address, count int) bool {
	// Perform transfers from Alice to Bob (1 token each)
	for i := 0; i < count; i++ {
		if err := transferTokens(env, ALICE_ADDRESS, BOB_ADDRESS, contract, TRANSFER_AMOUNT); err != nil {
			fmt.Printf("❌ Transfer %d failed: %v\n", i+1, err)
			return false
		}
	}
	return true
}

func formatTokenAmount(amount *big.Int) string {
	// Convert from wei to tokens (divide by 10^18)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	tokens := new(big.Int).Div(amount, divisor)
	return tokens.String()
} 