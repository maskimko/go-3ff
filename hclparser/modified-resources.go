package hclparser

//ModifiedResources class holds the data about modifications were made. It is a mutable struct
type ModifiedResources struct {
	items, authors map[string]struct{}
}

func (mr *ModifiedResources) Add(s string) {
	if _, ok := mr.items[s]; !ok {
		mr.items[s] = struct{}{}
	}
}

//List function lists the modified resources from the ModifiedResources object
func (mr *ModifiedResources) List() *[]string {
	s := make([]string, 0, len(mr.items))
	for k := range mr.items {
		s = append(s, k)
	}
	return &s
}

//IsEmpty function returns true if there were no modifications
func (mr *ModifiedResources) IsEmpty() bool {
	return len(mr.items) == 0
}

//AddAuthor function adds author s to the map of modification authors
func (mr *ModifiedResources) AddAuthor(s string) {
	if _, ok := mr.authors[s]; !ok {
		mr.authors[s] = struct{}{}
	}
}

//ListAuthors shows the list of modification authors
func (mr *ModifiedResources) ListAuthors() *[]string {
	s := make([]string, 0, len(mr.authors))
	for k := range mr.authors {
		s = append(s, k)
	}
	return &s
}

//NewModifiedResources function creates a new instance of ModifiedResources object
func NewModifiedResources() *ModifiedResources {
	return &ModifiedResources{
		items:   make(map[string]struct{}),
		authors: make(map[string]struct{}),
	}
}
