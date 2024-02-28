
### Callee 

A callee is a unique identifier for a function that can be invoked by an external transaction. A callee is uniquely identified by a combination of the contract's address and the function signature.

The Scheduler have two functions:

1. `Add(lftAddr [20]byte, lftSig [4]byte, rgtAddr [20]byte, rgtSig [4]byte)`: Add a new conflict pair to the scheduler.
This function should be called to update its internal conflict db after the detection. Duplicate pairs will be ignored.

1. `New(stdMsgs []*eucommon.StandardMessage)`: Create a schedule based on the input messages. The scheduler will return a schedule object based on:
    - The input messages
    - The conflict history

### Schedule
The Schedule is struct containing different types of transactions.

```go
	-Transfers    []*eucommon.StandardMessage // Transfers
	-Deployments  []*eucommon.StandardMessage // Contract deployments
	-Unknows      []*eucommon.StandardMessage // Messages with unknown conflicts with others
	-WithConflict []*eucommon.StandardMessage // Messages with conflicts
	-Sequentials  []*eucommon.StandardMessage // Callees that are marked as sequential only
	-Generations  [][]*eucommon.StandardMessage
	-CallCounts   []map[string]int
```
### Execution Order
There two types of execution orders: `Sequential` and `Parallel`.  

#### Sequential
Sequential transcactions are the ones to be execution in a single thread. 
- Transfers
- WithConflict
- Sequentials

#### Parallel
Parallel transactions are the ones that can be executed in parallel.
- Unknows
- Generations
- Deployments
