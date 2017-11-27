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
	"sync"
)

/*

STM ::

-------

Represents the STM (Software Transactional Memory). It is the only piece of
shared memory in the framework.

* `size`: It's vestigial in this case. However, if we were using C, it could have been of great importance.
The `size` represents the size of memory to be used by the STM.

* `_Memory`: It's the vector that holds the `MemoryCell`s.

* `memMutex`: It's the mutual exclusive lock for _Memory

* `_Ownerships`: It's the vector that holds the `Ownership` objects.

* `ownerMutex`: It's the mutex lock for synchronizing `_Ownerships`

* `transactions`: It's the vector that holds the pointers to `Transaction`s.

* `tMutex`: It's the mutex lock for synchronizing `transactions`

*/
type STM struct {
	//size        uint
	_Memory      []*MemoryCell
	memMutex     *sync.Mutex // mutex for synchronizing `_Memory`
	_Ownerships  []*Ownership
	ownerMutex   *sync.Mutex // mutex for synchronizing `_Ownerships`
	transactions []*Transaction
	tMutex       *sync.Mutex // mutex for synchronizing `transactions`
}

/*
Ownership ::

--------------

`Ownership` is a structure that links the memory cells to it's owning transaction (`Record`).
It has pointers to both the `MemoryCell` and the owner `Record`.

* `memoryCell`: The pointer to the `MemoryCell`.

* `owner`: the pointer to the `Transaction` or the owner transaction.

*/
type Ownership struct {
	memoryCell *MemoryCell
	owner      *Transaction
}

/*
NewSTM :: Creates a new STM instance. This acts as the single shared space.
*/
func NewSTM() *STM {
	stm := new(STM)
	stm._Memory = make([]*MemoryCell, 5)
	stm.memMutex = new(sync.Mutex)
	stm._Ownerships = make([]*Ownership, 5)
	stm.ownerMutex = new(sync.Mutex)
	stm.transactions = make([]*Transaction, 5)
	stm.tMutex = new(sync.Mutex)
	return stm
}

/*
MakeMemCell :: Makes a new `MemoryCell` holding the data.
*/
func (stm *STM) MakeMemCell(data *Data) *MemoryCell {
	newMemCell := new(MemoryCell)
	newMemCell.cellIndex = uint(len(stm._Memory))
	newMemCell.data = data
	//# synchronized memory cell creation
	stm.memMutex.Lock()
	stm._Memory = append(stm._Memory, newMemCell)
	stm.memMutex.Unlock()
	//# synchronized memory cell creation
	return newMemCell
}
