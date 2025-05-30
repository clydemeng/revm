// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title SimpleERC20
 * @dev A simple ERC20 token implementation for benchmarking
 */
contract SimpleERC20 {
    string public name = "BenchmarkToken";
    string public symbol = "BENCH";
    uint8 public decimals = 18;
    uint256 public totalSupply;
    
    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;
    
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    
    constructor(uint256 _totalSupply) {
        totalSupply = _totalSupply;
        balanceOf[msg.sender] = _totalSupply;
        emit Transfer(address(0), msg.sender, _totalSupply);
    }
    
    function transfer(address to, uint256 value) public returns (bool) {
        require(balanceOf[msg.sender] >= value, "Insufficient balance");
        balanceOf[msg.sender] -= value;
        balanceOf[to] += value;
        emit Transfer(msg.sender, to, value);
        return true;
    }
    
    function transferFrom(address from, address to, uint256 value) public returns (bool) {
        require(balanceOf[from] >= value, "Insufficient balance");
        require(allowance[from][msg.sender] >= value, "Insufficient allowance");
        
        balanceOf[from] -= value;
        balanceOf[to] += value;
        allowance[from][msg.sender] -= value;
        
        emit Transfer(from, to, value);
        return true;
    }
    
    function approve(address spender, uint256 value) public returns (bool) {
        allowance[msg.sender][spender] = value;
        emit Approval(msg.sender, spender, value);
        return true;
    }
}

/**
 * @title TransferHelper
 * @dev Helper contract for performing batch ERC20 transfers efficiently
 */
contract TransferHelper {
    SimpleERC20 public token;
    
    event BatchTransferCompleted(uint256 count, uint256 totalAmount);
    event SingleTransfer(address indexed from, address indexed to, uint256 amount);
    
    constructor(address _token) {
        token = SimpleERC20(_token);
    }
    
    /**
     * @dev Perform a single ERC20 transfer
     * @param from Source address
     * @param to Destination address  
     * @param amount Amount to transfer
     */
    function performTransfer(address from, address to, uint256 amount) public returns (bool) {
        // First approve this contract to spend tokens
        // Note: In real scenario, this would be done separately
        require(token.transferFrom(from, to, amount), "Transfer failed");
        emit SingleTransfer(from, to, amount);
        return true;
    }
    
    /**
     * @dev Perform multiple transfers in a single transaction
     * @param from Source address
     * @param recipients Array of recipient addresses
     * @param amounts Array of amounts to transfer
     */
    function batchTransfer(
        address from,
        address[] calldata recipients,
        uint256[] calldata amounts
    ) public returns (bool) {
        require(recipients.length == amounts.length, "Array length mismatch");
        
        uint256 totalAmount = 0;
        for (uint256 i = 0; i < recipients.length; i++) {
            require(token.transferFrom(from, recipients[i], amounts[i]), "Transfer failed");
            totalAmount += amounts[i];
            emit SingleTransfer(from, recipients[i], amounts[i]);
        }
        
        emit BatchTransferCompleted(recipients.length, totalAmount);
        return true;
    }
    
    /**
     * @dev Perform n identical transfers for benchmarking
     * @param from Source address
     * @param to Destination address
     * @param amount Amount per transfer
     * @param count Number of transfers to perform
     */
    function benchmarkTransfers(
        address from,
        address to,
        uint256 amount,
        uint256 count
    ) public returns (bool) {
        for (uint256 i = 0; i < count; i++) {
            require(token.transferFrom(from, to, amount), "Transfer failed");
            emit SingleTransfer(from, to, amount);
        }
        
        emit BatchTransferCompleted(count, amount * count);
        return true;
    }
    
    /**
     * @dev Get token balance of an address
     */
    function getBalance(address account) public view returns (uint256) {
        return token.balanceOf(account);
    }
    
    /**
     * @dev Get token allowance
     */
    function getAllowance(address owner, address spender) public view returns (uint256) {
        return token.allowance(owner, spender);
    }
} 