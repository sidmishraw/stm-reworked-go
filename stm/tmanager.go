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

/*

STM ::

-------

Represents the STM (Software Transactional Memory). It is the only piece of
shared memory in the framework.

* `size`: It's vestigial in this case. However, if we were using C, it could have been of great importance.
The `size` represents the size of memory to be used by the STM.

* `_Memory`: It's the vector that holds the `MemoryCell`s.

* `_Ownerships`: It's the vector that holds the `Ownership` objects.

* `records`: It's the vector that holds the `Record` objects that represent the transaction's metadata.

*/
type STM struct {
	//size        uint
	_Memory     []*MemoryCell
	_Ownerships []*Ownership
	records     []*Record
}

/*
Ownership ::

--------------

`Ownership` is a structure that links the memory cells to it's owning transaction (`Record`).
It has pointers to both the `MemoryCell` and the owner `Record`.

* `memoryCell`: The pointer to the `MemoryCell`.

* `record`: the pointer to the `Record` or the owner transaction.

*/
type Ownership struct {
	memoryCell *MemoryCell
	record     *Record
}

/*

MakeMemCell ::

----------------

Makes a new `MemoryCell` holding the data.

@param {*Data} data The pointer to the Data
@return {*MemoryCell} The pointer to the MemoryCell
*/
func (stm *STM) MakeMemCell(data *Data) *MemoryCell {
	newMemCell := new(MemoryCell)
	newMemCell.cellIndex = uint(len(stm._Memory))
	newMemCell.data = data
	stm._Memory = append(stm._Memory, newMemCell)
	return newMemCell
}
