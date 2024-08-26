
# Rollback
- [Rollback](#rollback)
  - [1. What Is Rollback?](#1-what-is-rollback)
  - [2. What Does Rollback Do?](#2-what-does-rollback-do)
  - [3. Why Use Rollback?](#3-why-use-rollback)
  - [4. Comparison with selfdestruct in Ethereum](#4-comparison-with-selfdestruct-in-ethereum)
  - [5. Implications of Rollback](#5-implications-of-rollback)
    - [5.1. Contract Address Conflict](#51-contract-address-conflict)
    - [5.2. An Example](#52-an-example)
  - [6. Solutions](#6-solutions)

## 1. What Is Rollback?

Rollback is a feature offered by Arcology Network that allows a contract to revert the state changes it made in the same block as when the `rollback` function is called. It is mainly designed to save gas fees spent on the concurrent contracts provided by the Arcology Network.

## 2. What Does Rollback Do?

The `rollback` function, when called, scans through the storage snapshot and undos all the state changes belonging to the contract that called the `rollback` function. This effectively reverts the state of the contract to the last block. All the changes made up to the point in the current block of the calling function call are removed.

## 3. Why Use Rollback?

Arcology Network provides a set of tools helping developers create concurrent contracts that can fully utilize the parallel processing capabilities of the network. For better modularity, these tools are provided in the form of a set of contracts that can be imported into the main contract. 

These contracts can provide significant flexibility and power to developers, but they come at a cost. Although some of these tools are merely utility functions and do not inherently necessitate any storage for their operation, using them will still consume gas for storage. Althrough the gas fees are dramatically lower on Arcology Network than on Ethereum, it is still a cost that should be avoided if possible.

The `rollback` feature is designed to cut the cost on unnecessary storage caused by deploying these utility contracts. It allows the contracts to revert the state changes it made in the same block, which can save a lot of gas fees.

## 4. Comparison with selfdestruct in Ethereum

The `selfdestruct` function in Ethereum allows a contract to destroy itself and send its remaining balance to another address. The `rollback` function in Arcology Network shares some similarities with `selfdestruct` in Ethereum, but there are some key differences.

| Feature | Rollback | selfdestruct |
| --- | --- | --- |
| Destroy the contract | No / Yes(Only when called in deployment block) | Yes |
| Send the remaining balance to | Message sender | A specified address |
|Computation gas refund | No | No |
|Storage gas refund | Full in the deployment block, partial elsewhere | Partial |

The `rollback` function can only be called in the same block as the deployment of the contract. This means that the contract cannot be destroyed by calling `rollback` after the block is mined.

## 5. Implications of Rollback

Rollback is a powerful feature that can save a lot of gas fees, but there are some deep implications that should be considered. The single biggest implication is that the `rollback` function can cause address conflict in deployment. In Ethereum The deployment address of the contract is determined by:

1. The deployer's address
2. The deployer's nonce

If both the deployer's address and the nonce are the same, the deployment address will be the same. This isn't a problem in Ethereum, because the nonce is incremented by 1 for each transaction sent by the same address. Calling `selfdestruct` will not destroy the contract completely. Let alone its
address and nonce.

### 5.1. Contract Address Conflict

The problem occurs when the rollback function of a `deployer contract` is called in the same block as the contract has just deployed another contract. In this case, the `rollback` function will revert the state of the `deployer contract` to the last block, including the nonce value of the contract. However, later when the deployer contract tries to deploy another contract, the nonce it uses will be the same as the one used in the previous deployment, causing the newly deployed contract to have the same address as the previous one.

### 5.2. An Example

Now consider the following the example:

- The contract `deployer` is first called by a transcaction to deploy a contract `B`. Before it, the `deployer` contract has an internal nonce value of `10` from the last block. After the deployment, the `deployer` contract increments its nonce to `11`. 

<img width="500" src="./img/first-deployment.png">
  
- Another transaction **in the same block** is calling `deployer` to invoke the `rollback` function. The `rollback` function will revert the state of the contract `deployer` back to the last block, which includes the nonce value. which will be `10` again. 
  
<img width="330" src="./img/rollback.png">

- In a later block, when the `deployer` contract trys to deploy another contract `C`, the nonce it uses will be 10, which is the same as the one used to deploy the contract `B` in the previous deployment. This will cause the the contract `C` to have the same address as the contract `B`. 

<img width="500" src="./img/second-deployment.png">

## 6. Solutions

There are a number of solutions to the address conflict problem:

1. A contract should not be able to use `rollback` if it has deployed another contract in the same block.

2. A contract should be able to use `rollback` only in same block it is deployed.

3. When `rollback` is called, revert everything but the nonce.

4. Use a deterministic random nonce for each contract-initiated deployment.
