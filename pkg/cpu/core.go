package cpu

type Core struct {
	pc uint32
	x  [32]uint32
}

func NewCore() *Core {
	return &Core{
		pc: 0,
		x:  [32]uint32{},
	}
}

func (c *Core) SetPc(value uint32) {
	c.pc = value
}

func (c *Core) GetPc() uint32 {
	return c.pc
}
