package m3gl

type SizeVar struct {
	Min float64
	Max float64
	Val float64
}

func (v SizeVar) getDelta() float64 {
	return (v.Val - v.Min) / (10.0 * (v.Max - v.Min))
}

func (v *SizeVar) check() {
	if v.Val >= v.Max {
		v.Val = v.Max
	}
	if v.Val <= v.Min {
		v.Val = v.Min
	}
}

func (v *SizeVar) Increase() {
	v.Val += (v.Max - v.Val) / 10.0
	v.check()
}

func (v *SizeVar) Decrease() {
	v.Val -= (v.Val - v.Min) / 10.0
	v.check()
}
