package inf

// SimulatedStack represents a stack using a slice of integers.
type SimulatedStack struct {
	items []int
}

// NewSimulatedStack creates and returns a new SimulatedStack.
func NewSimulatedStack() *SimulatedStack {
	return &SimulatedStack{items: []int{}}
}

// PushBack adds an item to the end of the stack.
func (s *SimulatedStack) PushBack(item int) {
	s.items = append(s.items, item)
}

// Back returns the last item of the stack without removing it.
// Returns an error if the stack is empty.
func (s *SimulatedStack) Back() int {
	if len(s.items) == 0 {
		panic("stack is empty")
	}
	return s.items[len(s.items)-1]
}

// PopBack removes and returns the last item from the stack.
// Returns an error if the stack is empty.
func (s *SimulatedStack) PopBack() int {
	if len(s.items) == 0 {
		panic("stack is empty")
	}
	index := len(s.items) - 1
	item := s.items[index]
	s.items = s.items[:index]
	return item
}

func (s *SimulatedStack) Size() int {
	return len(s.items)
}
