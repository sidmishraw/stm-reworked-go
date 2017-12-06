/**
* tmanager.go
* @author Sidharth Mishra
* @description The STM or transaction manager. It is the only piece of shared
memory in this framework.
* @created Wed Nov 22 2017 21:59:44 GMT-0800 (PST)
* @copyright 2017 Sidharth Mishra
* @last-modified Wed Nov 22 2017 23:56:05 GMT-0800 (PST)
*/

package stm

import (
	"bytes"
	"log"
	"sync"
)

/*

STM ::

-------

Represents the STM (Software Transactional Memory). It is the only piece of
shared memory in the framework.

* `_Memory`: It's the vector that holds the `MemoryCell`s.

* `_Ownerships`: It's the vector that holds the MemoryCell's ownerships

*/
type STM struct {
	stmMutex    *sync.Mutex          // stm's mutex
	_Memory     []*MemoryCell        // MemoryCells
	_Ownerships map[int]*Transaction // *Ownership
}

/*
@Deprecated

Ownership ::

--------------

`Ownership` is a structure that links the memory cells to it's owning transaction (`Record`).
It has pointers to both the `MemoryCell` and the owner `Record`.

* `memoryCell`: The pointer to the `MemoryCell`.

* `owner`: the pointer to the `Transaction` or the owner transaction.

*/
// type Ownership struct {
// 	memoryCell *MemoryCell
// 	owner      *Transaction
// }

/*
NewSTM :: Creates a new STM instance. This acts as the single shared space.
*/
func NewSTM() *STM {
	stm := new(STM)
	stm.stmMutex = new(sync.Mutex)
	stm._Memory = make([]*MemoryCell, 0)
	stm._Ownerships = make(map[int]*Transaction, 0)
	return stm
}

/*
MakeMemCell :: Makes a new `MemoryCell` holding the data.
*/
func (stm *STM) MakeMemCell(data Data) *MemoryCell {
	newMemCell := new(MemoryCell)
	newMemCell.cellIndex = uint(len(stm._Memory))
	newMemCell.data = new(bytes.Buffer)       // make the bytes buffer to hold the memcell's data
	writeStatus := newMemCell.writeData(data) // encode the bytes and write into the buffer
	if !writeStatus {
		log.Fatalln("Failed to create MemoryCell - Encoding failed!")
	}
	//# add memory cell to STM - synchoronously
	stm.stmMutex.Lock()
	stm._Memory = append(stm._Memory, newMemCell)
	stm.stmMutex.Unlock()
	//# add memory cell to STM - synchoronously
	return newMemCell
}

/*
Exec :: Executes the transactions and holds the calling thread so that it doesn't exit prematurely.
This is just an utility method to make life easier for the consumer. The consumer can also use
Transaction's Go() to achieve this, but then the consumer has to pass their own sync.WaitGroup instance.

> Note: This just shortens the code written, it does have the same effect as the following piece of code
 		wg := new(sync.WaitGroup)
		wg.Add(2)
		t1.Go(wg)
		t2.Go(wg)
		wg.Wait()

> Note: Make sure that the STM instance executing the transaction is same as the one which was used to
construct it. Otherwise, it will result in an error since the shared memory won't be the same.
*/
func (stm *STM) Exec(ts ...*Transaction) {
	wg := new(sync.WaitGroup)
	for _, t := range ts {
		wg.Add(1)
		t.Go(wg)
	}
	wg.Wait()
}

// Display displays the _Memory array of the STM
func (stm *STM) Display() {
	for i, memcell := range stm._Memory {
		log.Println("memcell index = ", i, " memcell contents = ", *memcell.data)
	}
	log.Println("_Ownerships = ", stm._Ownerships)
}

// ForkAndExec forks from the calling thread and then executes all the transactions on the
// forked thread. Basically it can be visualized as running the stm.Exec in another thread.
// The consumer can simulate similar behavior by doing something like:
// `go MySTM.Exec(ts...)`. This is a convinience method to keep it uniform.
func (stm *STM) ForkAndExec(ts ...*Transaction) {
	go func(ts ...*Transaction) {
		wg := new(sync.WaitGroup)
		for _, t := range ts {
			wg.Add(1)
			t.Go(wg)
		}
		wg.Wait()
	}(ts...)
}

// Log logs the messages synchronously
func (stm *STM) Log(msgs ...interface{}) {
	log.Println(msgs...)
}
