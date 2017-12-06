/**
* memory-cells.go
* @author Sidharth Mishra
* @description Definitions of MemoryCell and its related methods/functions.
* @created Wed Nov 22 2017 21:44:55 GMT-0800 (PST)
* @copyright 2017 Sidharth Mishra
* @last-modified Thu Nov 23 2017 18:37:01 GMT-0800 (PST)
 */

package stm

import (
	"bytes"
	"encoding/gob"
)

/*

MemoryCell ::

-------------

Represents each memory cell that holds data.

* `cellIndex`: The index or address of the `MemoryCell` in the `_Memory` vector of the STM.
// to be used internally

* `data`: The data stored inside the `MemoryCell`. It is a byte buffer.

*/
type MemoryCell struct {
	cellIndex uint
	data      *bytes.Buffer
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
// func (cell *MemoryCell) GetData() Data {
// 	return *cell.data
// }

// writeData writes the bytes into the byte buffer of the MemoryCell
// usage:
// status := memCell.writeData(Data([]int{1,2,3}))
// if !status { log.Fataln("Faield to write!")}
func (memCell *MemoryCell) writeData(data Data) bool {
	encoder := gob.NewEncoder(memCell.data)
	memCell.data.Reset() // reset the buffer before new data writes
	err := encoder.Encode(data)
	if nil != err {
		// failed to encode data
		return false
	}
	return true
}

// readData reads the contents of the MemoryCell into the dataContainer
// usage:
// m := make([]int, 0)
// status := memCell.readData(Data(m))
// if !status {
//  log.Fataln("Failed to read data out of the MemoryCell")
// }
func (memCell *MemoryCell) readData(dataContainer Data) bool {
	encoder := gob.NewEncoder(memCell.data)
	decoder := gob.NewDecoder(memCell.data)
	err := decoder.Decode(dataContainer) // decode the data into the dataContainer
	if err != nil {
		return false
	}
	err = encoder.Encode(dataContainer) // re-encode a copy of the data before moving on - to maintain the data in the buffer for future reads
	if err != nil {
		return false
	}
	return true
}
