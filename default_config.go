package fuzzy_pid

const (
	NB = -3 + iota
	NM
	NS
	ZO
	PS
	PM
	PB
)

var defaultKpMatrix = [][]int{
	{PB, PB, PM, PM, PS, ZO, ZO},
	{PB, PB, PM, PS, PS, ZO, NS},
	{PM, PM, PM, PS, ZO, NS, NS},
	{PM, PM, PS, ZO, NS, NM, NM},
	{PS, PS, ZO, NS, NS, NM, NM},
	{PS, ZO, NS, NM, NM, NM, NB},
	{ZO, ZO, NM, NM, NM, NB, NB},
}

var defaultKiMatrix = [][]int{
	{NB, NB, NM, NM, NS, ZO, ZO},
	{NB, NB, NM, NS, NS, ZO, ZO},
	{NB, NM, NS, NS, ZO, PS, PS},
	{NM, NM, NS, ZO, PS, PM, PM},
	{NM, NS, ZO, PS, PS, PM, PB},
	{ZO, ZO, PS, PS, PM, PB, PB},
	{ZO, ZO, PS, PM, PM, PB, PB},
}

var defaultKdMatrix = [][]int{
	{PS, NS, NB, NB, NB, NM, PS},
	{PS, NS, NB, NM, NM, NS, ZO},
	{ZO, NS, NM, NM, NS, NS, ZO},
	{ZO, NS, NS, NS, NS, NS, ZO},
	{ZO, ZO, ZO, ZO, ZO, ZO, ZO},
	{PB, NS, PS, PS, PS, PS, PB},
	{PB, PM, PM, PM, PS, PS, PB},
}

var eMfParas = []float64{-3, -3, -2, -3, -2, -1, -2, -1, 0, -1, 0, 1, 0, 1, 2, 1, 2, 3, 2, 3, 3}
var deMfParas = []float64{-3, -3, -2, -3, -2, -1, -2, -1, 0, -1, 0, 1, 0, 1, 2, 1, 2, 3, 2, 3, 3}
var KpMfParas = []float64{-3, -3, -2, -3, -2, -1, -2, -1, 0, -1, 0, 1, 0, 1, 2, 1, 2, 3, 2, 3, 3}
var KiMfParas = []float64{-3, -3, -2, -3, -2, -1, -2, -1, 0, -1, 0, 1, 0, 1, 2, 1, 2, 3, 2, 3, 3}
var KdMfParas = []float64{-3, -3, -2, -3, -2, -1, -2, -1, 0, -1, 0, 1, 0, 1, 2, 1, 2, 3, 2, 3, 3}

func (pid *FuzzyPid) SetDefaultMf() *FuzzyPid {
	pid.SetMF(
		Trimf, eMfParas,
		Trimf, deMfParas,
		Trimf, KpMfParas,
		Trimf, KiMfParas,
		Trimf, KdMfParas,
	)
	return pid
}

func (pid *FuzzyPid) SetDefaultRuleMatrix() *FuzzyPid {
	pid.SetRuleMatrix(defaultKpMatrix, defaultKiMatrix, defaultKdMatrix)
	return pid
}

func (pid *FuzzyPid) SetDefaultConfig() *FuzzyPid {
	pid.SetDefaultMf()
	pid.SetDefaultRuleMatrix()
	return pid
}
