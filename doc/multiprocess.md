
<!-- # 1. Multiprocessor -->
- [1. What is the MP](#1-what-is-the-mp)
- [2. Using the MP](#2-using-the-mp)
  - [2.1. Adding to the Job Queue](#21-adding-to-the-job-queue)
  - [2.2. Running the Parallel Jobs](#22-running-the-parallel-jobs)
  - [2.3. Failed Transactions](#23-failed-transactions)
  - [2.4. Clearing the Queue](#24-clearing-the-queue)
  - [2.5. Rollbacking the MP State](#25-rollbacking-the-mp-state)
  - [2.6. Example](#26-example)
  - [2.7. Nested MP](#27-nested-mp)
- [3. Behind the Scene](#3-behind-the-scene)
  - [3.1. Workflows](#31-workflows)
  - [3.2. Snapshot](#32-snapshot)
  - [3.2. Conflicts \& Merging](#32-conflicts--merging)
- [4. Implications and Solutions](#4-implications-and-solutions)
  - [4.1. Problems with Parallel Deployment](#41-problems-with-parallel-deployment)
  - [4.2 Solutions](#42-solutions)
  - [4.2. Example](#42-example)

The multiprocessor is a great tool for creating thread-like sub transactions in Solidty. It is allows contracts to process multiple sub transactions in parallel. It is espicallly useful for compuationally intensive tasks. 

## 1. What is the MP

The MP is a contract that can create sub transactions. The MP doesn't change the way the EVM works. So, it seems logical to see MP == EOA in the sense that they both can initiate transactions. However, it is not the case. The MP is still a contract, and it cannot pay the gas fee. The MP is more like a sub transaction creation proxy that can create sub transactions on behalf of the caller. The caller is the one who owns the sub transactions. It is the message sender that pays the gas fee. 

## 2. Using the MP

To use the MP, the first step is to create an instance of the MP. The constructor of the MP takes the concurrency level as an argument. The concurrency level is the number of sub transactions that can be processed in parallel. They number can take any value between 1 and 256. Anything above 256 will be capped at 256. It the number of sub transactions is greater than the concurrency level, the sub transactions will be processed in batches. 

### 2.1. Adding to the Job Queue

Once the MP is created, you can add jobs to the MP. The jobs are added to the MP using the `push()` function. The `push()` function takes the gas limit, the target address, and the data as arguments. The data is the function call that you want to make. It has very similar syntax to the Ethereum `call()` function.

### 2.2. Running the Parallel Jobs

After all the jobs are added to the queue, ths MP will start processing the jobs in parallel once the function `run()` is called, using the number of threads specified in the constructor. The clear state changes will be merged together and updated in the main thread. 

###  2.3. Failed Transactions

A MP created sub transaction can failed for two reasons:

- Execution failure: The sub transaction fails to execute just like a standard transaction. 

- State Conflict: This happens when executions for multiple sub transactions are successful, but the final state changes cannot be merged together without removing some transaction that are causing the conflicts. 

### 2.4. Clearing the Queue

The last step is to call `clear()` to clear the queue. It is important to clear the queue after the jobs are processed. Otherwise, the jobs will be processed again when the `run()` function is called again.

### 2.5. Rollbacking the MP State

As described above, the MP is more like a proxy contract spawning the sub transactions. Once it is instantized, there is still a need to write some state in the storage, incurring gas costs. So, when a MP instance is no longer needed, it is always recommended to call `rollback()` to clear it up. The function will remove all the storage occupied of the MP instance. A full refund of all the gas spent on storage changes happend in the current block will be given back to the transaction sender. If the contract has received funds in the current block before the function is called, those funds will also be sent back to the message sender too.

### 2.6. Example

In the example below, the MP is used to process two sub transactions in parallel with two threads. The first sub transaction calls the `assigner()` function with the arguments 0 and 10. The second sub transaction calls the `assigner()` function with the arguments 1 and 20. The `assigner()` function assigns the value to the array at the index specified by the first argument.

```solidity
pragma solidity >= 0.8.0 < 0.9.0;
import "@arcologynetwork/concurrent/lib/mulitprocess/Multiprocess.sol";

contract ParallelAssigner {
    uint256[2] array; 

    constructor() { 
       Multiprocess mp = new Multiprocess(2); 
       mp.push(5000000, address(this), abi.encodeWithSignature("assigner(uint256,uint256)", 0, 10));
       mp.push(5000000, address(this), abi.encodeWithSignature("assigner(uint256,uint256)", 1, 20));
       mp.run();
    }

    function assign(uint256 idx, uint256 v) public {
        array[idx] = v
    }

    require(array[0] == 10);
    require(array[1] == 20);
} 
```

###  2.7. Nested MP

Recursive MP calls are possible but constrained by the depth and the concurrency level. The depth is the number of levels of nested MP calls. The depth limit is 4. Maximizing parallelism by using nested MPs isn't recommended. It is better to use a higher concurrency level for single MP calls. 

```solidity
pragma solidity >= 0.8.0 < 0.9.0;
import "@arcologynetwork/concurrent/lib/mulitprocess/Multiprocess.sol";

contract NestedAssigner {
    uint256[4] array; 

    function call() public { 
       Multiprocess mp = new Multiprocess(2); 
       mp.push(5000000, address(this), abi.encodeWithSignature("proxy(uint256)", 0));
       mp.push(5000000, address(this), abi.encodeWithSignature("proxy(uint256)", 1));
       mp.run();

       require(array[0] == 10);
       require(array[1] == 21);   
       require(array[2] == 12);
       require(array[3] == 23); 
    }

    function proxy(uint256 idx) public {
       Multiprocess mp = new Multiprocess(2); 
       mp.push(2500000, address(this), abi.encodeWithSignature("assign(uint256,uint256,uint256)", idx, 0, 10));
       mp.push(2500000, address(this), abi.encodeWithSignature("assign(uint256,uint256,uint256)", idx, 1, 20));
       mp.run();        
    }

    function assign(uint256 idx, uint256 i, uint256 v) public {
        array[idx*2 + i] = v + idx*2 + i;
    }   
} 
```

## 3. Behind the Scene

When initiated, the MP is nothing more than a container to temporarily store the necessary information for creating sub transactions. Parallel jobs can be then added to the MP for further processing.  The real magic happens when the `run()` function is called. At that point, the MP will create sub transactions and create EVM instance for them. The EVMs will process the sub transactions in parallel.

<img width="700" src="./img/snapshot.png">


### 3.1. Workflows 

When the `run` is called, the MP does the following:

- It creates a snapshot for every sub transaction. The snapshot is a copy of the current state up to the point that the function is called.
  
- Then, it initializes a number of EVMs based on the concurrency level specified in the constructor. 
  
- It convert the function call info stored in the MP to transactions and send them to the EVMs.
  
- The EVMs will process the sub transactions in parallel again its own snapshot.
  
- Once the sub transactions are processed, the MP will collect the state changes of the sub transactions and scan for conflicts. 
  
- Tthe MP will revert the state changed of the sub transactions that are causing the conflicts. 

- Finally, the MP will merge the clean state changes of the sub transactions back to the main thread.

### 3.2. Snapshot

The snapshot is a copy of the current state up to the point that the function is called. It is used to create a fresh state for the sub transactions. The snapshot is created based on the current snapshot. When multiple sub transactions are created in parallel, they all their own independent snapshots. Which means all the data written to the storage by the sub transactions are not visible to each other.


### 3.2. Conflicts & Merging

Conflicts occur when different sub transactions are trying to write/read to the same storage location. The MP will scan for conflicts after the sub transactions are processed. The MP will revert the state changed of the sub transactions that are causing the conflicts.

In the below example, the MP is used to process two sub transactions in parallel with two threads. The first sub transaction calls the `assign()` function with the an argument of 1. The second sub transaction calls the `assign()` function with an argument of 2. Because the two sub transactions are trying to write to the same variable, the MP will revert the state changes of the second sub transaction.

MP determines which transaction to revert based on a number of factors. But the process is deterministic and predictable.

```solidity
pragma solidity >= 0.8.0 < 0.9.0;
import "@arcologynetwork/concurrent/lib/mulitprocess/Multiprocess.sol";

contract SimpleConflict {
    uint256 data;
    function call() public {
        Multiprocess mp = new Multiprocess(2);
        mp.push(100000, address(this), abi.encodeWithSignature("assign(uint256)", 1)); // Only one will go through
        mp.push(100000, address(this), abi.encodeWithSignature("assign(uint256)", 2)); // Only one will go through
        mp.run();
        require(data == 1);
    }

    function assign(uint256 v) public { 
        data = v;
    } 
}
```

## 4. Implications and Solutions

MP is a very powerful tool that can save a lot of gas fees, but it doesn't come without its implications. Thera are a few issues that need to be addressed. 

### 4.1. Problems with Parallel Deployment

One of the main issues with deploying multiple contracts in parallel is the potential for conflicts. This occurs because, when deploying contracts in multiprocessor-created transactions. Independent snapshots are created based on the current snapshot right before the sub transactions are processed and there is no inter-thread communication mechanism during the execution.

This doesn't cause any problem if the sub transactions are not deploying contracts. However, if they are deploying contracts, they will deploy them to the same address. This is because:

- The initial nonce of the caller address is the same for all transactions, and as the nonce is ncremented by 1 for each transaction, they all end up with the **same caller nonce**.
  
- The target contract address is calculated as based on caller address + nonce. Because the caller address and nonce are the same for all transactions, the target contract address is the same for all transactions. Consequently, all these transaction are all going to conflict with each other.


### 4.2 Solutions

There are a few solutions to this problem. 

1. The simplest solution is banning the deployment of contracts in the MP. This is not a good solution because it limits the functionality of the MP. The MP should be able to deploy contracts. 

2. The second solution is not to care about the conflicts. If the happen, they happen. This, again, is not a good solution because it will limit the functionality of the MP.

3. The third solution is to chang the rules for deploying contracts. Currently, a contract's deployment address is based on the caller's address and nonce. It is possible to decouple the deployment address from the caller's address and nonce. This has significant implications, not only for the MP but also for the entire Ethereum network. It is not a viable solution."

4. Another solution is to disperse the nonce value for each MP created sub transaction. This can be done either by offsetting the initial nonce or by incrementing the nonce by a pseudo-random number. This way, the transactions are deployed to different addresses.

For more information, see [Address Conflict](./nonce.md).


