/**
* memory-cells.go
* @author Sidharth Mishra
* @description Definitions of MemoryCell and its related methods/functions.
* @created Wed Nov 22 2017 21:44:55 GMT-0800 (PST)
* @copyright 2017 Sidharth Mishra
* @last-modified Thu Nov 23 2017 18:37:01 GMT-0800 (PST)
 */

package stm

// MemoryCell represents each memory cell that holds data.
// `cellIndex`: The index or address of the `MemoryCell` in the `_Memory` vector of the STM.
// to be used internally
// `data`: The data stored inside the `MemoryCell`
type MemoryCell struct {
	cellIndex uint
	data      Data
}

// writeData writes the bytes into the byte buffer of the MemoryCell
// usage:
// status := memCell.writeData(Data([]int{1,2,3}))
// if !status { log.Fataln("Faield to write!")}
func (memCell *MemoryCell) writeData(data Data) {
	memCell.data = data
}

// readData reads the contents of the MemoryCell into the dataContainer
// new Usage:
// data := memCell.readData() // of type Data
func (memCell *MemoryCell) readData() Data {
	return memCell.data.Clone()
}
