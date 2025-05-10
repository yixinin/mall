package set

type Set[T comparable] struct {
	m map[T]struct{}
}

func New[T comparable](cap ...int) *Set[T] {
	if len(cap) == 0 {
		cap = append(cap, 0)
	}
	return &Set[T]{
		m: make(map[T]struct{}, cap[0]),
	}
}

func From[T comparable](ks []T) *Set[T] {
	s := &Set[T]{
		m: make(map[T]struct{}, len(ks)),
	}
	s.Add(ks...)
	return s
}

func (s *Set[T]) Add(k ...T) *Set[T] {
	for _, k := range k {
		s.m[k] = struct{}{}
	}
	return s
}

func (s *Set[T]) Del(k ...T) *Set[T] {
	for _, k := range k {
		delete(s.m, k)
	}
	return s
}

func (s *Set[T]) Has(k T) bool {
	_, ok := s.m[k]
	return ok
}

func (s *Set[T]) ToSlice() []T {
	var ks = make([]T, 0, len(s.m))
	for k := range s.m {
		ks = append(ks, k)
	}
	return ks
}
