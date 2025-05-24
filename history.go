package repllib

type ReplHistory interface {
	Push(buffer string) error
	GetAll() ([]string, error)
}

// Simple in-memory history implementation
type memoryHistory struct {
	commands []string
}

func (h *memoryHistory) Push(buffer string) error {
	h.commands = append(h.commands, buffer)
	return nil
}

func (h *memoryHistory) GetAll() ([]string, error) {
	return h.commands, nil
}
