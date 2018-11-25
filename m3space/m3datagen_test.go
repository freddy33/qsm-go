package m3space

import (
	"fmt"
	"gonum.org/v1/gonum/stat"
	"testing"
)

func TestStatPack(t *testing.T) {
	fmt.Println(stat.StdDev([]float64{1.3,1.5,1.7,1.1}, nil))
}
