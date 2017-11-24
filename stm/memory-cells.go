/**
* memory-cells.go
* @author Sidharth Mishra
* @description Definitions of MemoryCell and its related methods/functions.
* @created Wed Nov 22 2017 21:44:55 GMT-0800 (PST)
* @copyright 2017 Sidharth Mishra
* @last-modified Thu Nov 23 2017 18:37:01 GMT-0800 (PST)
 */

package stm

/*

MemoryCell ::

-------------

Represents each memory cell that holds data.

* `cellIndex`: The index or address of the `MemoryCell` in the `_Memory` vector of the STM.
// to be used internally

* `data`: The data stored inside the `MemoryCell`.

*/
type MemoryCell struct {
	cellIndex uint
	data      *Data
}

/*

GetData ::

---------

Fetches the data contained in the `MemoryCell`.
It just returns the value of the data contained in the MemoryCell. It doesn't give the
reference. This is to make sure that the consumer doesn't accidentally update the MemoryCell's
contents.

@return {Data} the data contained in the `MemoryCell`
*/
func (cell *MemoryCell) GetData() Data {
	return *cell.data
}
