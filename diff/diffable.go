//diff package contains functions for computation diff of HCL data structures
package diff

//Diffable interface enforces DiffParam fuction to be present in order to compute difference
type Diffable interface {
	//DiffParam function returns string to consider for comparison operations
	DiffParam() string
}
