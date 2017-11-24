/**
* transaction.go
* @author Sidharth Mishra
* @description Contains definitions of the `Record` object.
* @created Wed Nov 22 2017 21:59:31 GMT-0800 (PST)
* @copyright 2017 Sidharth Mishra
* @last-modified Thu Nov 23 2017 18:08:05 GMT-0800 (PST)
 */

package stm

/*
Record ::

-----------

Represents a record that contains the metadata for a transaction.

* `status` - the status of the transaction, true if it has sucessfully completed, else false

* `version` - the version of the transaction, initially starts at 0, its incremented after every termination

* `oldValues` - the vector containing the old values of the memory cells when they are updated in the transaction.

* `readSet` - the set of names - addresses - of the memory cells that the transaction intends to read from.

* `writeSet` - the set of names - addresses - of the memory cells that the transaction intends to write into or update.
*/
type Record struct {
	status    bool
	version   int
	oldValues []Data
	readSet   []string
	writeSet  []string
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

@return {*TransactionContext} the new transaction context
*/
func (stm *STM) NewT() *TransactionContext {
	tc := new(TransactionContext)
	tc.transaction = new(Transaction)
	tc.actions = make([]func() bool, 5)
	return tc
}

/*
Do :: Is used to chain actions together in the transaction.
Each action is wrapped inside an anonymous lambda. This method will add a wrapper and pass the
`Transaction` so that the action is within the desired Transaction's context.
In DB terms, this will represent an `Action`. A transaction comprises of multiple `Actions`.

@param {func() bool} lambda The anonymous function that signifies an action
@return {*TransactionContext} the updated transaction context
*/
func (tc *TransactionContext) Do(lambda func(*Transaction) bool) *TransactionContext {
	tc.actions = append(tc.actions, func() bool { return lambda(tc.transaction) })
	return tc
}

/*
Done :: Gets the componentized `Transaction` that can be passed around and used as needed.

@return {*Transaction} the pointer to the componentized `Transaction`
*/
func (tc *TransactionContext) Done() *Transaction {
	tc.transaction.metadata = Record{
		status:    false,
		version:   0,
		oldValues: make([]Data, 5),
		readSet:   make([]string, 5),
		writeSet:  make([]string, 5),
	}
	tc.transaction.actions = tc.actions
	return tc.transaction
}
