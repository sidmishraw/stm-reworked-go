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
