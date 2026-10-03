package outcome

//sumtype:decl
//nolint:iface // a sealed union is switched on only outside its package, and its marker method only seals it.
type Outcome interface {
	isOutcome()
}
