## eu
The EU project introduces an Abstract Execution Unit that serves as a transaction processing unit on the Arcology network. This module is designed to be VM-agnostic, providing a versatile solution for transaction processing. The module comprises two fundamental components.

- **[Parallelized EVM](https://github.com/arcology-network/concurrent-evm):** A module responsible for parallelizing the Ethereum Virtual Machine (EVM) on Arcology Network.

- **[EVM-adaptor](https://github.com/arcology-network/vm-adaptor):** A module functioning as a middleware to connect to the parallelized EVM, managing executable messages as input and producing state transitions as output.

<p align="center">
<img src="/eu.png" alt="eu" width="924">
</p>

## Input and Output

- **Input:** Executable messages from either the executor module or the Multiprocessor API calls.

- **Output:** State transitions generated as output from the EVM-adaptor module.


## Multiprocessor Handling

The Multiprocessor module within the EU project enables developers to manually initiate multiple independent threads for concurrent function execution on the Arcology network using the [concurrent API in Solidity](https://github.com/arcology-network/concurrentlib). It is the focal point for handling all multiprocessor-related logic within the EU project. The module is responsible for the following tasks:

- Start Multiprocessor:
Initiates the Multiprocessor to enable parallel processing.

- Wrap Function Calls into Messages:
Converts function calls into messages for concurrent processing.

- Start Multiple EVM Instances:
Initiates multiple instances of the Ethereum Virtual Machine.

- Make Snapshots from Individual Threads:
Captures snapshots from individual threads for further processing.

- Feed Messages with State Snapshot and Start Processing:
Combines messages with state snapshots and initiates parallel processing.

- Handling Runtime Errors:
Manages errors occurring during execution at runtime.

- Logging Execution-Related Errors:
Logs errors encountered during the execution process.

- Manage and Terminate Threads:
Handles the creation and termination of threads initiated by the Multiprocessor.

- Revert Transition Changes:
Reverts transition changes generated by threads in case of errors.

- Merge State Changes by Different Threads:
Merges state changes from different threads back into the main thread.

- Merge State Changes Back to the Main Thread:
Finalizes the process by merging state changes back into the main thread for a coherent result.

- Terminate the Threads:
The threads will be terminated when all executions have been completed, together with thire state snapshots.

## Usage
For details on how to integrate and use the EU project's Multiprocessor module, refer to the documentation.
Feel free to contribute and report issues in the GitHub repository.

## License
This project is licensed under the MIT License.