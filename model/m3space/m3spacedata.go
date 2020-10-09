package m3space

import (
	"github.com/freddy33/qsm-go/m3util"
)

var Log = m3util.NewLogger("m3space", m3util.INFO)

type SpacePackDataIfc interface {
	m3util.QsmDataPack
	GetAllSpaces() []SpaceIfc
	GetSpace(id int) SpaceIfc
	CreateSpace(name string, activePathNodeThreshold DistAndTime,
		maxTriosPerPoint int, maxPathNodesPerPoint int) (SpaceIfc, error)
	DeleteSpace(id int, name string) (int, error)
}

type BaseSpacePackData struct {
	EnvId m3util.QsmEnvID
}

func (ppd *BaseSpacePackData) GetEnvId() m3util.QsmEnvID {
	if ppd == nil {
		return m3util.NoEnv
	}
	return ppd.EnvId
}

func CreateAllIndexes(nbIndexes int) ([][4]int, [12]int) {
	// TODO: Equivalence not evident at all finally dues to relation between trioIdx and the axis
	// the points of the pyramid are equivalent. So, any reorder that end up with same array of 4 is equivalent.
	// so creating the arrays starting from previous index will create the combinations of indexes

	// This are not true combinations as each index can be duplicated
	// So, its the sum of t4 (all the same) + t3 (3 identical) + t2a (2 identical) + t2b ( 2 x 2 identical) + t1 (all different)
	// Formula | for nb = 8 | for nb = 12
	// t4 = nbIndexes | 8 | 12
	// t3 = nbIndexes * (nbIndexes-1) | 56 | 132
	// t2a = nbIndexes * (nbIndexes-1) / 2 | 28 | 66
	// t2b = nbIndexes * ( (nbIndexes-1)! / ( 2! (nbIndexes-1-2)! ) ) | 168 | 660
	// with n is number of indexes and k=4 we have number of combinations: t1 = n! / ( k! (n-k)! ) | 70 | 495
	var t4, t3, t2a, t2b, t1 int
	if nbIndexes == 8 {
		t4 = 8
		t3 = 56
		t2a = 28
		t2b = 168
		t1 = 70
	} else if nbIndexes == 12 {
		t4 = 12
		t3 = 132
		t2a = 66
		t2b = 660
		t1 = 495
	} else {
		Log.Fatalf("Nb indexes %d is not supported", nbIndexes)
		return nil, [12]int{}
	}
	nbConbinations := t4 + t3 + t2a + t2b + t1
	res := make([][4]int, nbConbinations)
	idx := 0
	var nbT4, nbT3, nbT2a, nbT2b, nbT1 int
	for i1 := 0; i1 < nbIndexes; i1++ {
		res[idx] = [4]int{i1, i1, i1, i1}
		idx++
		nbT4++
		for iT3 := 0; iT3 < nbIndexes; iT3++ {
			if iT3 != i1 {
				res[idx] = [4]int{i1, i1, i1, iT3}
				idx++
				nbT3++
			}
		}
		for i2 := i1 + 1; i2 < nbIndexes; i2++ {
			res[idx] = [4]int{i1, i1, i2, i2}
			idx++
			nbT2a++
			for iT2b := 0; iT2b < nbIndexes; iT2b++ {
				if iT2b != i1 && iT2b != i2 {
					res[idx] = [4]int{i1, i2, iT2b, iT2b}
					idx++
					nbT2b++
				}
			}
			for i3 := i2 + 1; i3 < nbIndexes; i3++ {
				for i4 := i3 + 1; i4 < nbIndexes; i4++ {
					res[idx] = [4]int{i1, i2, i3, i4}
					idx++
					nbT1++
				}
			}
		}
	}
	return res, [12]int{nbConbinations, idx, t1, nbT1, t2a, nbT2a, t2b, nbT2b, t3, nbT3, t4, nbT4}
}

