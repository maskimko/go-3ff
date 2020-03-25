package hclparser

type ModifiedResources struct {
	items, authors map[string]struct{}
}

func (mr *ModifiedResources) Add(s string) {
	if _, ok := mr.items[s]; !ok {
		mr.items[s] = struct{}{}
	}
}

func (mr *ModifiedResources) List() *[]string {
	s := make([]string, 0, len(mr.items))
	for k := range mr.items {
		s = append(s, k)
	}
	return &s
}

func (mr *ModifiedResources) IsEmpty() bool {
	return len(mr.items) == 0
}

func (mr *ModifiedResources) AddAuthor(s string) {
	if _, ok := mr.authors[s]; !ok {
		mr.authors[s] = struct{}{}
	}
}

func (mr *ModifiedResources) ListAuthors() *[]string {
	s := make([]string, 0, len(mr.authors))
	for k := range mr.authors {
		s = append(s, k)
	}
	return &s
}

func NewModifiedResources() *ModifiedResources {
	return &ModifiedResources{
		items:   make(map[string]struct{}),
		authors: make(map[string]struct{}),
	}
}
