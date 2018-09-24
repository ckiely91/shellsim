package command

type CommandHistory struct {
	history []string
	idx     int
}

func (c *CommandHistory) Append(line string) {
	c.history = append(c.history, line)
	c.idx = len(c.history)
}

func (c *CommandHistory) Up() string {
	if len(c.history) == 0 {
		return ""
	}
	if c.idx > 0 {
		c.idx--
	}
	return c.history[c.idx]
}

func (c *CommandHistory) Down() string {
	if c.idx >= len(c.history) {
		return ""
	}
	c.idx++
	if c.idx >= len(c.history) {
		return ""
	}
	return c.history[c.idx]
}
