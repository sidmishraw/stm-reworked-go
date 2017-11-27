/**
* utils.go
* @author Sidharth Mishra
* @description Utility functions
* @created Fri Nov 24 2017 23:28:15 GMT-0800 (PST)
* @copyright 2017 Sidharth Mishra
* @last-modified Fri Nov 24 2017 23:28:23 GMT-0800 (PST)
 */

package stm

/* contains - Checks if the data is in the container */
func contains(container []uint, data uint) bool {
	for _, value := range container {
		if value == data {
			return true
		}
	}
	return false
}

/* remove - Removes the data from the container */
func remove(container []uint, data uint) []uint {
	index := -1
	for i, v := range container {
		if v == data {
			index = i
		}
	}
	if index != -1 {
		return append(container[0:index], container[index+1:]...)
	}
	return container // does nothing and passes the container unmodified
}

/*
alreadyOwned :: Returns true if the MemoryCell is already owned by some other transaction else returns false
*/
func alreadyOwned(ownerships []*Ownership, ownership *Ownership) bool {
	for _, o := range ownerships {
		if o != ownership && o.memoryCell == ownership.memoryCell && o.owner != ownership.owner {
			return true
		}
	}
	return false
}

/*
isTheOwner :: Checks if the current Transaction is the owner of the MemoryCell, else returns false
*/
func isTheOwner(ownerships []*Ownership, ownership *Ownership) bool {
	for _, o := range ownerships {
		if o.memoryCell == ownership.memoryCell && o.owner == ownership.owner {
			return true
		}
	}
	return false
}

/* releaseOwnership - Releases the ownership of the MemoryCell */
func releaseOwnership(ownerships []*Ownership, ownership *Ownership) []*Ownership {
	index := -1
	for i, v := range ownerships {
		if v.memoryCell == ownership.memoryCell && v.owner == ownership.owner {
			index = i
		}
	}
	if index != -1 {
		return append(ownerships[0:index], ownerships[index+1:]...) // removes the ownership record from the ownerships vector
	}
	return ownerships // does nothing and passes the ownerships unmodified
}
