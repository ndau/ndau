# Generate minter.go package
1. Compile the latest Minter.sol contract with:
    `solc --abi --base-path '/' --include-path 'node_modules/' contracts/L1/Minter.sol`
2. Copy the abi JSON of the Minter contract only and paste to a file named Minter.abi
3. Generate minter.go package from abi file
    `abigen --abi=./Minter.abi --pkg=minter > ./minter.go`