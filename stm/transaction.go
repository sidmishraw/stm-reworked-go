/**
* transaction.go
* @author Sidharth Mishra
* @description Contains definitions of the `Record` object.
* @created Wed Nov 22 2017 21:59:31 GMT-0800 (PST)
* @copyright 2017 Sidharth Mishra
* @last-modified Thu Nov 23 2017 18:08:05 GMT-0800 (PST)
 */

package stm

import (
	"reflect"
	"sync"
)

/*
Record ::

-----------

Represents a record that contains the metadata for a transaction.

* `status` - the status of the transaction, true if it has sucessfully completed, else false

* `version` - the version of the transaction, initially starts at 0, its incremented after every termination

* `oldValues` - the vector containing the old values of the memory cells when they are updated in the transaction.

* `readSet` - the set of MemoryCell indices - addresses - of the memory cells that the transaction intends to read from.

* `writeSet` - the set of MemoryCell indices - addresses - of the memory cells that the transaction intends to write to or update.
*/
type Record struct {
	status    bool
	version   int
	oldValues map[*MemoryCell]Data
	readSet   []*MemoryCell
	writeSet  []*MemoryCell
}

/*
Transaction ::

---------------

The transaction, as a component. This can be passed around. It has its own context.
It will carry out the actions mentioned while constructing it and will always be consistent.
*/
type Transaction struct {
	metadata Record
	actions  []func() bool
	stm      *STM
}

/*
TransactionContext ::

----------------------

A transaction context - just like Monads - inspired by Monads.

This defines a context within which the STM operations will be available.
*/
type TransactionContext struct {
	actions     []func() bool
	transaction *Transaction
}

/*
NewT :: Makes a new transaction context.
*/
func (stm *STM) NewT() *TransactionContext {
	tc := new(TransactionContext)
	tc.transaction = new(Transaction)
	tc.actions = make([]func() bool, 0)
	tc.transaction.stm = stm
	//# synchronized addition of transactions to shared STM
	stm.tMutex.Lock()
	stm.transactions = append(stm.transactions, tc.transaction)
	stm.tMutex.Unlock()
	//# synchronized addition of transactions to shared STM
	return tc
}

/*
Do :: Is used to chain actions together in the transaction.
Each action is wrapped inside an anonymous lambda. This method will add a wrapper and pass the
`Transaction` so that the action is within the desired Transaction's context.
In DB terms, this will represent an `Action`. A transaction comprises of multiple `Actions`.
*/
func (tc *TransactionContext) Do(lambda func(*Transaction) bool) *TransactionContext {
	tc.actions = append(tc.actions, func() bool { return lambda(tc.transaction) })
	return tc
}

/*
Done :: Gets the componentized `Transaction` that can be passed around and used as needed.
*/
func (tc *TransactionContext) Done() *Transaction {
	tc.transaction.metadata = Record{
		status:    false,
		version:   0,
		oldValues: make(map[*MemoryCell]Data, 0),
		readSet:   make([]*MemoryCell, 0),
		writeSet:  make([]*MemoryCell, 0),
	}
	tc.transaction.actions = tc.actions
	return tc.transaction
}

/*
ReadT :: A transactional read operation. Reads the data from the passed MemoryCell instance.
When reading a MemoryCell, the trasaction doesn't need to take ownership.
*/
func (t *Transaction) ReadT(memcell *MemoryCell) Data {
	//# Adding to read set
	// If the address of the memory cell is not in the writeset
	// then, add it into the ReadSet, else do nothing
	if !contains(t.metadata.writeSet, memcell) && !contains(t.metadata.readSet, memcell) {
		t.metadata.readSet = append(t.metadata.readSet, memcell)
	}
	//# Adding to read set
	data := memcell.data
	// take backup into the oldValues
	//# backup
	t.metadata.oldValues[memcell] = *data
	//# backup
	return *data
}

/*
WriteT :: A transactional write/update operation. Writes the data into the MemoryCell.
When intending to write to a MemoryCell, a transaction must take ownership of the MemoryCell.
If the transaction failed to take ownership of the MemoryCell, write fails. Returns true when the data
is successfully written into the MemoryCell.
*/
func (t *Transaction) WriteT(memcell *MemoryCell, data Data) (succeeded bool) {
	//# Adding to write set
	if contains(t.metadata.readSet, memcell) {
		t.metadata.readSet = remove(t.metadata.readSet, memcell)
	}
	if !contains(t.metadata.writeSet, memcell) {
		t.metadata.writeSet = append(t.metadata.writeSet, memcell)
	}
	//# Adding to write set
	//# Take ownership of the memCell and write
	tOwn := new(Ownership)
	tOwn.memoryCell = memcell
	tOwn.owner = t
	if !alreadyOwned(t.stm._Ownerships, tOwn) {
		// since the MemoryCell is not owned by any Transactions, take ownership before
		// the Write operation.
		//# synchronized ownership acquired
		t.stm.ownerMutex.Lock()
		t.stm._Ownerships = append(t.stm._Ownerships, tOwn)
		t.stm.ownerMutex.Unlock()
		//# synchronized ownership acquired
		currentData := memcell.data
		//#backup
		t.metadata.oldValues[memcell] = *currentData
		//#backup
		memcell.data = &data // data updated
		succeeded = true
	} else if isTheOwner(t.stm._Ownerships, tOwn) {
		// already the owner of the MemoryCell so no need to take ownership again
		// proceed with the Write operation.
		currentData := memcell.data
		//#backup
		t.metadata.oldValues[memcell] = *currentData
		//#backup
		memcell.data = &data // data updated
		succeeded = true
	} else {
		succeeded = false
	}
	//# Take ownership of the memCell and write
	return succeeded
}

/*
Go :: Starts executing the `Transaction t`.
Keeps looping infinitely, retrying the actions of the transaction until it executes successfully.
*/
func (t *Transaction) Go(wg *sync.WaitGroup) {
	//# spawn and execute in new thread/goroutine
	go func() {
		//# Transaction's execution loop, keeps retrying till it successfully executes
		for {
			if exStatus := t.executeActions(); !exStatus {
				// execute all the actions for the Transaction t, upon success exStatus = true else false
				// rollback the transaction since the actions have failed to execute successfully
				// fmt.Println("t execution failed, rollback")
				t.rollback()
				continue
			}
			if cmtStatus := t.commit(); cmtStatus {
				// the actions of the transaction have executed successfully
				// and the commit operation was successful
				t.metadata.status = true // updating the status to true signifying that the transaction executed successfully
				t.metadata.version++     // updating the version signifying successful end of the transaction
				break
			} else {
				// the actions of the transaction executed properly, but,
				// the commit operation failed, so, rollback and continue the transaction
				// from the beginning.
				// fmt.Println("t commit failed, rollback")
				t.rollback()
				continue
			}
		}
		//# Transaction's execution loop, keeps retrying till it successfully executes
		wg.Done()
	}()
	//# spawn and execute in new thread/goroutine
}

/*
executeActions :: Executes the actions serially, returns true if all the actions were executed successfully, else returns false.
*/
func (t *Transaction) executeActions() bool {
	for _, action := range t.actions {
		status := action()
		if !status {
			return false
		}
	}
	return true
}

/*
rollback :: Rollsback the `Transaction t`. This includes restoring the oldValues of all the write set members
*/
func (t *Transaction) rollback() {
	// to rollback the transaction, restore the backups in the
	// Transaction's metadata called oldValues.
	for _, wsMemCell := range t.metadata.writeSet {
		backup := t.metadata.oldValues[wsMemCell] // fetch backup from the map
		//# synchronized updation of _Memory
		t.stm.memMutex.Lock()
		t.stm._Memory[wsMemCell.cellIndex].data = &backup // restore the backup
		t.stm.memMutex.Unlock()
		//# synchronized updation of _Memory
		//# release ownership
		dummyTOwn := new(Ownership)
		dummyTOwn.memoryCell = t.stm._Memory[wsMemCell.cellIndex]
		dummyTOwn.owner = t
		//# synchronized release of ownership
		t.stm.ownerMutex.Lock()
		t.stm._Ownerships = releaseOwnership(t.stm._Ownerships, dummyTOwn)
		t.stm.ownerMutex.Unlock()
		//# synchronized release of ownership
		//# release ownership
	}
}

/*
commit :: Commits the Transaction t. After committing, the Transaction releases the ownership of the MemoryCells and their values become visible to all the other transactions.
Commit depends on the readSet members. If the value of the readSet members have changed in the meantime,
the commit should fail and the Transaction should rollback and restart from the beginning.
The commit failure is signified by a `cmtStatus = false`. The success is represented as `cmtStatus = true`.
*/
func (t *Transaction) commit() (cmtStatus bool) {
	cmtStatus = true // let's assume we have a successful commit
	for _, rsMemCell := range t.metadata.readSet {
		backup := t.metadata.oldValues[rsMemCell] // get the Transaction's backup to compare against the current state in STM
		current := *t.stm._Memory[rsMemCell.cellIndex].data
		if !reflect.DeepEqual(backup, current) {
			// since the backup and current values don't match
			// there might be a modification and the this Transaction's
			// computation might be wrong now, need to rollback and retry
			cmtStatus = false
			break
		}
	}
	//# release ownership of MemoryCells in the write set
	// only release ownership in case of successful commit
	// otherwise the ownership will be released by the rollback subroutine
	if cmtStatus {
		for _, wsMemCell := range t.metadata.writeSet {
			//# release ownership
			dummyTOwn := new(Ownership)
			dummyTOwn.memoryCell = t.stm._Memory[wsMemCell.cellIndex]
			dummyTOwn.owner = t
			//# synchronized release of ownership
			t.stm.ownerMutex.Lock()
			t.stm._Ownerships = releaseOwnership(t.stm._Ownerships, dummyTOwn)
			t.stm.ownerMutex.Unlock()
			//# synchronized release of ownership
			//# release ownership
		}
	}
	//# release ownership of MemoryCells in the write set
	return cmtStatus
}
