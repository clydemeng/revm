// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract MinimalERC20 {
    mapping(address => uint256) public balanceOf;
    uint256 public totalSupply;
    
    constructor() {
        totalSupply = 1000000 * 10**18; // 1 million tokens
        balanceOf[msg.sender] = totalSupply;
    }
    
    function transfer(address to, uint256 amount) public returns (bool) {
        require(balanceOf[msg.sender] >= amount, "Insufficient balance");
        balanceOf[msg.sender] -= amount;
        balanceOf[to] += amount;
        return true;
    }
} 