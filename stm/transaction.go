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
	"bytes"
	"encoding/gob"
	"log"
	"reflect"
	"sync"
)

//# For Debugging

// use the flag to denote that all logs in transactions should be printed to stdout/stderr
const debugFlag bool = false

//# For Debugging

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
	name      string
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
	metadata   Record
	actions    []func() bool
	stm        *STM
	isScanning bool            // true value indicates that the transaction is in Scan mode
	tvars      map[string]Data // map of all the transactional variables
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
	tc.transaction.tvars = make(map[string]Data, 0)
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

// Where function is used to add transactional variables. This allows the transaction's
// actions to be pure to some extent.
// usage:
// MySTM.NewT().
// 	Where("foo", stm.Data([]int{1,2,3,4,5})).
// 	Do(func(t *stm.Transaction) bool {...}).
// 	Done()
func (tc *TransactionContext) Where(name string, value Data) *TransactionContext {
	tc.transaction.tvars[name] = value // add the props
	return tc
}

/*
Done :: Gets the componentized `Transaction` that can be passed around and used as needed.
*/
func (tc *TransactionContext) Done(name ...string) *Transaction {
	tName := "t"
	if len(name) != 0 {
		tName = name[0]
	}
	tc.transaction.metadata = Record{
		name:      tName,
		status:    false,
		version:   0,
		oldValues: make(map[*MemoryCell]Data, 0),
		readSet:   make([]*MemoryCell, 0),
		writeSet:  make([]*MemoryCell, 0),
	}
	tc.transaction.actions = tc.actions
	tc.transaction.isScanning = true // default is true
	return tc.transaction
}

//# Transactional Variables

// GetTVar gets the value of the Transactional Variable for the name provided.
// Note: It will return `nil` when the transaction is in its `Scan Phase`. Make sure to
// check for `nil` values.
// Usage:
// v := t.GetTVar("foo")
// if v != nil {
//   // this portion of the code will not run in scan phase
//	// so make sure that you don't put `WriteT` operations inside this
// }
func (t *Transaction) GetTVar(name string) Data {
	if t.isScanning {
		return nil
	}
	return t.tvars[name]
}

// PutTVar puts the value into Transactional variable of the given name.
// Note: It will do nothing when the transaction is in its `Scan Phase`.
// Usage:
// t.PutTVar("foo", []int{1,2,3})
func (t *Transaction) PutTVar(name string, value Data) {
	if t.isScanning {
		return
	}
	t.tvars[name] = value
}

//# Transactional Variables

/*
ReadT :: A transactional read operation. Reads the data from the passed MemoryCell instance.
When reading a MemoryCell, the trasaction doesn't need to take ownership.
*/
func (t *Transaction) ReadT(memcell *MemoryCell, dataContainer Data) bool {
	//# read data from stm
	t.stm.stmMutex.Lock()
	rStatus := t.stm._Memory[memcell.cellIndex].readData(dataContainer)
	t.stm.stmMutex.Unlock()
	if !rStatus {
		return false
	}
	//# read data from stm
	//# Adding to read set
	// If the address of the memory cell is not in the writeset
	// then, add it into the ReadSet, else do nothing
	if t.isScanning {
		if !contains(t.metadata.writeSet, memcell) && !contains(t.metadata.readSet, memcell) {
			t.metadata.readSet = append(t.metadata.readSet, memcell)
		}
		t.log(t.metadata.name, " Scanning, added ", memcell, " to readSet", t.metadata.readSet)
		t.log(t.metadata.name, " and got data = ", dataContainer)
		return true // early return, no need to take backup during scan phase
	}
	//# Adding to read set
	//# backup
	// take backup into the oldValues
	t.metadata.oldValues[memcell] = dataContainer
	//# backup
	return true
}

/*
WriteT :: A transactional write/update operation. Writes the data into the MemoryCell.
When intending to write to a MemoryCell, a transaction must take ownership of the MemoryCell.
If the transaction failed to take ownership of the MemoryCell, write fails. Returns true when the data
is successfully written into the MemoryCell.
*/
func (t *Transaction) WriteT(memcell *MemoryCell, data Data) (succeeded bool) {
	//# Adding to write set
	if t.isScanning {
		// if contains(t.metadata.readSet, memcell) {
		// 	t.metadata.readSet = remove(t.metadata.readSet, memcell)
		// }
		if !contains(t.metadata.writeSet, memcell) {
			t.metadata.writeSet = append(t.metadata.writeSet, memcell)
		}
		t.log(t.metadata.name, " Scanning, added ", memcell, " to writeSet", t.metadata.writeSet)
		t.log(t.metadata.name, " readSet = ", t.metadata.readSet)
		return true // no need to write the contents into the memorycell during scan phase
	}
	//# Adding to write set
	//# Check ownership of the memCell and write to oldValues
	t.stm.stmMutex.Lock()
	owner := t.stm._Ownerships[int(memcell.cellIndex)]
	t.stm.stmMutex.Unlock()
	if owner == t {
		// already the owner of the MemoryCell so no need to take ownership again
		// proceed with the Write operation.
		//# newData is stored in oldValues
		t.metadata.oldValues[memcell] = data
		//# newData is stored in oldValues
		succeeded = true
		t.log(t.metadata.name, "  already has ownership of ", memcell, " hence write was successful")
	} else {
		succeeded = false
		t.log(t.metadata.name, "  couldn't take ownership of ", memcell, " write operation has failed.")
	}
	//# Check ownership of the memCell and write to oldValues
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
			//# Scanning phase
			t.metadata.status = false // signal that t transaction has started execution
			t.log(t.metadata.name, "has started scanning")
			t.scanActions() // scan the actions to determine readSet and writeSet
			t.log(t.metadata.name, "has finished scanning")
			//# Scanning phase
			//# Ownerships phase
			t.log(t.metadata.name, "has started taking ownerships of writeSet members")
			if status := t.takeOwnerships(); !status {
				t.log(t.metadata.name, " has failed to take ownerships, rolling back and retrying")
				t.rollback()
				continue
			}
			t.log(t.metadata.name, "has taken ownerships of writeSet members")
			//# Ownerships phase
			//# Execution phase
			t.log(t.metadata.name, "has started execution")
			if exStatus := t.executeActions(); !exStatus {
				// execute all the actions for the Transaction t, upon success exStatus = true else false
				// rollback the transaction since the actions have failed to execute successfully
				t.rollback()
				t.log(t.metadata.name, " has failed to execute, rolling back and restarting")
				continue
			}
			t.log(t.metadata.name, "has finished execution")
			//# Execution phase
			//# Commit phase
			t.log(t.metadata.name, "has started commit phase")
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
				t.log(t.metadata.name, " has failed to commit, rolling back and restarting")
				t.rollback()
				continue
			}
			//# Commit phase
		}
		//# Transaction's execution loop, keeps retrying till it successfully executes
		t.log(t.metadata.name, " has successfully committed.")
		wg.Done()
	}()
	//# spawn and execute in new thread/goroutine
}

// scanActions scans the actions to determine
func (t *Transaction) scanActions() {
	t.isScanning = true // set the isScanning flag to true to signify that the scan has started
	for _, action := range t.actions {
		action() // execute the action in scan mode, don't bother about failing
	}
	t.isScanning = false // set the isScanning flag to false to signify that the scan has ended
}

// takeOwnerships signals the Ownership taking phase. If for some reason the transaction fails to take ownership, it will fail and repeat from the beginning
func (t *Transaction) takeOwnerships() bool {
	status := true // since there can be scenarios where there are no writeset members
	for _, wsMemCell := range t.metadata.writeSet {
		t.stm.stmMutex.Lock()
		owner := t.stm._Ownerships[int(wsMemCell.cellIndex)]
		t.stm.stmMutex.Unlock()
		if nil == owner {
			// since the MemoryCell is not owned by any Transactions, take ownership before
			//# synchronized ownership acquired
			t.stm.stmMutex.Lock()
			t.stm._Ownerships[int(wsMemCell.cellIndex)] = t
			t.stm.stmMutex.Unlock()
			//# synchronized ownership acquired
			status = true
			t.log(t.metadata.name, " has taken ownership of ", wsMemCell)
		} else if owner == t {
			// already the owner of the MemoryCell so no need to take ownership again
			status = true
			t.log(t.metadata.name, "  already has ownership of ", wsMemCell)
		} else {
			status = false
			t.log(t.metadata.name, "  couldn't take ownership of ", wsMemCell)
			break
		}
	}
	return status
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
rollback :: Rollsback the `Transaction t`.
Releasing the ownerships held by the transaction.
*/
func (t *Transaction) rollback() {
	// to rollback the transaction, restore the backups in the
	// Transaction's metadata called oldValues.
	for _, wsMemCell := range t.metadata.writeSet {
		//# release ownership
		t.stm.stmMutex.Lock()
		if t.stm._Ownerships[int(wsMemCell.cellIndex)] == t {
			t.stm._Ownerships[int(wsMemCell.cellIndex)] = nil // releases ownership
		}
		t.stm.stmMutex.Unlock()
		//# release ownership
	}
	//# reset the writeSet, readSet, and oldValues
	t.metadata.readSet = make([]*MemoryCell, 0)
	t.metadata.writeSet = make([]*MemoryCell, 0)
	t.metadata.oldValues = make(map[*MemoryCell]Data, 0)
	//# reset the writeSet, readSet, and oldValues
}

/*
commit :: Commits the Transaction t. After committing, the Transaction releases the ownership of the MemoryCells and their values become visible to all the other transactions.
Commit depends on the readSet members. If the value of the readSet members have changed in the meantime,
the commit should fail and the Transaction should rollback and restart from the beginning.
The commit failure is signified by a `cmtStatus = false`. The success is represented as `cmtStatus = true`.
*/
func (t *Transaction) commit() (cmtStatus bool) {
	cmtStatus = true // let's assume we have a successful commit
	//# check readSet members for inconsistencies
	for _, rsMemCell := range t.metadata.readSet {
		backup := t.metadata.oldValues[rsMemCell] // get the Transaction's backup to compare against the current state in STM
		t.stm.stmMutex.Lock()
		current := t.stm._Memory[rsMemCell.cellIndex].data
		t.stm.stmMutex.Unlock()
		buffer := new(bytes.Buffer)
		encoder := gob.NewEncoder(buffer)
		encoder.Encode(backup)
		t.log(t.metadata.name, "backup = ", backup, "and buffer bytes = ", buffer.Bytes(), "  and current bytes = ", current.Bytes())
		if !reflect.DeepEqual(buffer.Bytes(), current.Bytes()) && !contains(t.metadata.writeSet, rsMemCell) {
			// since the backup and current values don't match
			// there might be a modification and the this Transaction's
			// computation might be wrong now, need to rollback and retry
			t.log(t.metadata.name, "Readset member's Old and current values don't match -- failed")
			cmtStatus = false
			break
		}
	}
	//# check readSet members for inconsistencies
	//# release ownership of MemoryCells in the write set
	// only release ownership in case of successful commit
	// otherwise the ownership will be released by the rollback subroutine
	if cmtStatus {
		for _, wsMemCell := range t.metadata.writeSet {
			// first check if the write set members still hold the ownerships
			// otherwise fail the commit and let the transaction retry
			t.stm.stmMutex.Lock()
			owner := t.stm._Ownerships[int(wsMemCell.cellIndex)]
			t.stm.stmMutex.Unlock()
			if owner == t {
				// the write set member is still owned by the transaction
				// it is safe to write
				//# write new values to the memory location
				newData := t.metadata.oldValues[wsMemCell]
				t.log(t.metadata.name, "Preparing to write data into memcell, data = ", newData, " and memcell = ", wsMemCell)
				t.stm.stmMutex.Lock()
				t.stm._Memory[wsMemCell.cellIndex].writeData(newData) // write the new updated data
				//# write new values to the memory location
				//# release ownership
				//# synchronized release of ownership
				t.stm._Ownerships[int(wsMemCell.cellIndex)] = nil
				//# synchronized release of ownership
				//# release ownership
				t.stm.stmMutex.Unlock()
				t.log(t.metadata.name, "Wrote data into memcell, data = ", newData, " and memcell = ", wsMemCell)
			} else {
				// the writeset member is no longer held by the transaction
				// it is not safe to write, so the transaction should fail and retry
				cmtStatus = false // commit failed
				break             // break, need to fail the commit
			}
		}
	}
	//# release ownership of MemoryCells in the write set
	if cmtStatus {
		//# reset the writeSet, readSet, and oldValues
		t.metadata.readSet = make([]*MemoryCell, 0)
		t.metadata.writeSet = make([]*MemoryCell, 0)
		t.metadata.oldValues = make(map[*MemoryCell]Data, 0)
		//# reset the writeSet, readSet, and oldValues
	}
	return cmtStatus
}

//# For debugging

// GetVersion gets the version of the transaction
func (t *Transaction) GetVersion() int {
	return t.metadata.version
}

// log logs the messages to stderr/stdout depending on the debug flag
func (t *Transaction) log(msgs ...interface{}) {
	if debugFlag {
		log.Println(msgs...)
	}
}

//# For debugging
