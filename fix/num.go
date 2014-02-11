package fix

import (
	"math"
	"strconv"
)

const Precision = 20

type Num int64

func (n Num) Mul(m Num) Num {
	return (n * m) >> Precision
}

func (n Num) Div(m Num) Num {
	return (n << Precision) / m
}

func Int(i int) Num {
	return Num(i << Precision)
}

var floatScale = math.Pow(2, Precision)

func Float(f float64) Num {
	return Num(f * floatScale)
}

func (n Num) Float() float64 {
	return float64(n) / floatScale
}

func (n Num) String() string {
	return strconv.FormatFloat(n.Float(), 'f', -1, 64)
}
