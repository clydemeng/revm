// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract BIGA {
    string public name = "BIGA";
    string public symbol = "BIGA";
    uint8 public decimals = 18;
    uint256 public totalSupply;
    
    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;
    
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    
    constructor() {
        // Initialize with zero supply, use mint function to create tokens
    }
    
    function mint(address to, uint256 amount) public {
        totalSupply += amount;
        balanceOf[to] += amount;
        emit Transfer(address(0), to, amount);
    }
    
    function transfer(address to, uint256 amount) public returns (bool) {
        require(balanceOf[msg.sender] >= amount, "Insufficient balance");
        balanceOf[msg.sender] -= amount;
        balanceOf[to] += amount;
        emit Transfer(msg.sender, to, amount);
        return true;
    }
    
    function transferFrom(address from, address to, uint256 amount) public returns (bool) {
        require(balanceOf[from] >= amount, "Insufficient balance");
        require(allowance[from][msg.sender] >= amount, "Insufficient allowance");
        
        balanceOf[from] -= amount;
        balanceOf[to] += amount;
        allowance[from][msg.sender] -= amount;
        
        emit Transfer(from, to, amount);
        return true;
    }
    
    function approve(address spender, uint256 amount) public returns (bool) {
        allowance[msg.sender][spender] = amount;
        emit Approval(msg.sender, spender, amount);
        return true;
    }
    
    // Batch transfer function that performs multiple transfers in a single transaction
    function batchTransferSequential(address startRecipient, uint256 amountPerTransfer, uint256 numTransfers) public {
        require(balanceOf[msg.sender] >= amountPerTransfer * numTransfers, "Insufficient balance for batch transfer");
        
        uint256 recipientAddr = uint256(uint160(startRecipient));
        
        for (uint256 i = 0; i < numTransfers; i++) {
            address recipient = address(uint160(recipientAddr + i));
            balanceOf[msg.sender] -= amountPerTransfer;
            balanceOf[recipient] += amountPerTransfer;
            emit Transfer(msg.sender, recipient, amountPerTransfer);
        }
    }
} 