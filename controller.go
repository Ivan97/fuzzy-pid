package fuzzy_pid

import (
	"fmt"
	"math"
)

const N = 7

type FuzzyPid struct {
	target float64 //控制目标
	actual float64 //采样获得的真实值
	e      float64 //误差
	ePre1  float64
	ePre2  float64
	de     float64
	eMax   float64 //误差基本论域上限
	deMax  float64 //误差变化基本率论域上限

	deltaKpMax float64 //输出上限
	deltaKiMax float64
	deltaKdMax float64

	Ke  float64 //用于将e映射到论域[-3,3]上
	Kde float64 //用于将de映射到论域[-3,3]上
	KuP float64 //用于将△Kp结果映射到[-delta_Kp_max, delta_Kp_max]上
	KuI float64 //同理
	KuD float64 //同理

	KpRuleMatrix [N][N]float64 //Kp模糊规则矩阵
	KiRuleMatrix [N][N]float64 //Ki模糊规则矩阵
	KdRuleMatrix [N][N]float64 //Kd模糊规则矩阵

	mfTE  string //e的隶属函数类型
	mfTDe string //de的隶属函数类型
	mfTKp string //kp的隶属函数类型
	mfTKi string //ki的隶属函数类型
	mfTKd string //kd的隶属函数类型

	eMfParas  []float64 //e隶属函数参数
	deMfParas []float64 //de隶属函数参数
	kpMfParas []float64 //kp隶属函数参数
	kiMfParas []float64 //ki隶属函数参数
	kdMfParas []float64 //kd隶属函数参数

	Kp float64 //比例系数
	Ki float64 //积分系数
	Kd float64 //微分系数

	A float64
	B float64
	C float64
}

func NewFuzzyPid(eMax float64, deMax float64, kpMax float64, kiMax float64, kdMax float64, kp0 float64, ki0 float64, kd0 float64) *FuzzyPid {
	pid := &FuzzyPid{
		target: 0,
		actual: 0,

		eMax:       eMax,
		deMax:      deMax,
		deltaKpMax: kpMax,
		deltaKiMax: kiMax,
		deltaKdMax: kdMax,

		eMfParas:  nil,
		deMfParas: nil,
		kpMfParas: nil,
		kiMfParas: nil,
		kdMfParas: nil,

		e:     0,
		ePre1: 0,
		ePre2: 0,
		de:    0,
		Ke:    (N / 2) / eMax,
		Kde:   (N / 2) / deMax,
		KuP:   kpMax / (N / 2),
		KuI:   kiMax / (N / 2),
		KuD:   kdMax / (N / 2),

		mfTE:  "No Type",
		mfTDe: "No Type",
		mfTKp: "No Type",
		mfTKi: "No Type",
		mfTKd: "No Type",

		Kp: kp0,
		Ki: ki0,
		Kd: kd0,

		A: kp0 + ki0 + kd0,
		B: -2*kd0 - kp0,
		C: kd0,
	}
	return pid
}

func NewFuzzyPid_(fuzzyLimit []float64, pidInitVal []float64) *FuzzyPid {

	eMax := fuzzyLimit[0]
	deMax := fuzzyLimit[1]
	deltaKpMax := fuzzyLimit[2]
	deltaKiMax := fuzzyLimit[3]
	deltaKdMax := fuzzyLimit[4]

	Kp := pidInitVal[0]
	Ki := pidInitVal[1]
	Kd := pidInitVal[2]

	pid := &FuzzyPid{
		target: 0,
		actual: 0,

		eMax:       eMax,
		deMax:      deMax,
		deltaKpMax: deltaKpMax,
		deltaKiMax: deltaKiMax,
		deltaKdMax: deltaKdMax,

		eMfParas:  nil,
		deMfParas: nil,
		kpMfParas: nil,
		kiMfParas: nil,
		kdMfParas: nil,

		e:     0,
		ePre1: 0,
		ePre2: 0,
		de:    0,
		Ke:    (N / 2) / fuzzyLimit[0],
		Kde:   (N / 2) / fuzzyLimit[1],
		KuP:   deltaKpMax / (N / 2),
		KuI:   deltaKiMax / (N / 2),
		KuD:   deltaKdMax / (N / 2),

		mfTE:  "No Type",
		mfTDe: "No Type",
		mfTKp: "No Type",
		mfTKi: "No Type",
		mfTKd: "No Type",

		Kp: Kp,
		Ki: Ki,
		Kd: Kd,

		A: Kp + Ki + Kd,
		B: -2*Kd - Kp,
		C: Kd,
	}
	return pid
}

func (pid *FuzzyPid) GetKp() float64 {
	return pid.Kp
}

func (pid *FuzzyPid) GetKi() float64 {
	return pid.Ki
}

func (pid *FuzzyPid) GetKd() float64 {
	return pid.Kd
}

func (pid *FuzzyPid) GetA() float64 {
	return pid.A
}

func (pid *FuzzyPid) GetB() float64 {
	return pid.B
}

func (pid *FuzzyPid) GetC() float64 {
	return pid.C
}

func (pid *FuzzyPid) TrimF(x float64, a float64, b float64, c float64) float64 {
	var u float64

	if x >= a && x <= b {
		u = (x - a) / (b - a)
	} else if x > b && x <= c {
		u = (c - x) / (c - b)
	} else {
		u = 0
	}

	if u == math.NaN() {
		u = 0
	}

	return u
}

func (pid *FuzzyPid) SetRuleMatrix(kpM [][]int, kiM [][]int, kdM [][]int) {
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			pid.KpRuleMatrix[i][j] = float64(kpM[i][j])
			pid.KiRuleMatrix[i][j] = float64(kiM[i][j])
			pid.KdRuleMatrix[i][j] = float64(kdM[i][j])
		}
	}
}

func (pid *FuzzyPid) SetMFSub(_type string, paras []float64, n int) {

	NMfE := 0
	NMfDe := 0
	NMfKp := 0
	NMfKi := 0
	NMfKd := 0

	switch n {
	case 0:
		if _type == "trimf" || _type == "gaussmf" || _type == "trapmf" {
			pid.mfTE = _type
		} else {
			fmt.Println("Error: Type Error")
		}
		if pid.mfTE == "trimf" {
			NMfE = 3
		} else if pid.mfTE == "gaussmf" {
			NMfE = 2
		} else if pid.mfTE == "trapmf" {
			NMfE = 4
		}

		pid.eMfParas = make([]float64, N*NMfE)
		pid.eMfParas = paras

		break
	case 1:
		if _type == "trimf" || _type == "gaussmf" || _type == "trapmf" {
			pid.mfTDe = _type
		} else {
			fmt.Println("Error: Type Error")
		}
		if pid.mfTDe == "trimf" {
			NMfDe = 3
		} else if pid.mfTDe == "gaussmf" {
			NMfDe = 2
		} else if pid.mfTDe == "trapmf" {
			NMfDe = 4
		}
		pid.deMfParas = make([]float64, N*NMfDe)
		pid.deMfParas = paras
		break
	case 2:
		if _type == "trimf" || _type == "gaussmf" || _type == "trapmf" {
			pid.mfTKp = _type
		} else {
			fmt.Println("Error: Type Error")
		}
		switch pid.mfTKp {
		case "trimf":
			NMfKp = 3
			break
		case "gaussmf":
			NMfKp = 2
			break
		case "trapmf":
			NMfKp = 4
			break
		}
		pid.kpMfParas = make([]float64, N*NMfKp)
		pid.kpMfParas = paras
		break
	case 3:
		if _type == "trimf" || _type == "gaussmf" || _type == "trapmf" {
			pid.mfTKi = _type
		} else {
			fmt.Println("Error: Type Error")
		}
		switch pid.mfTKi {
		case "trimf":
			NMfKi = 3
			break
		case "gaussmf":
			NMfKi = 2
			break
		case "trapmf":
			NMfKi = 4
			break
		}
		pid.kiMfParas = make([]float64, N*NMfKi)
		pid.kiMfParas = paras
		break
	case 4:
		if _type == "trimf" || _type == "gaussmf" || _type == "trapmf" {
			pid.mfTKd = _type
		} else {
			fmt.Println("Error: Type Error")
		}
		switch pid.mfTKd {
		case "trimf":
			NMfKd = 3
			break
		case "gaussmf":
			NMfKd = 2
			break
		case "trapmf":
			NMfKd = 4
			break
		}
		pid.kdMfParas = make([]float64, N*NMfKd)
		pid.kdMfParas = paras
		break
	default:
		break
	}
}

func (pid *FuzzyPid) SetMF(
	mfTypeE string, emf []float64,
	mfTypeDe string, deMf []float64,
	mfTypeKp string, kpMf []float64,
	mfTypeKi string, kiMf []float64,
	mfTypeKd string, kdMf []float64) {
	pid.SetMFSub(mfTypeE, emf, 0)
	pid.SetMFSub(mfTypeDe, deMf, 1)
	pid.SetMFSub(mfTypeKi, kpMf, 2)
	pid.SetMFSub(mfTypeKp, kiMf, 3)
	pid.SetMFSub(mfTypeKd, kdMf, 4)
}

func (pid *FuzzyPid) Realize(t float64, a float64) float64 {
	var uE, uDe []float64
	var uEIndex, uDeIndex []int
	var deltaKp, deltaKi, deltaKd float64
	var deltaU float64

	uE = make([]float64, N)
	uDe = make([]float64, N)
	uEIndex = make([]int, 3)
	uDeIndex = make([]int, 3)

	pid.target = t
	pid.actual = a

	pid.e = pid.target - pid.actual
	pid.de = pid.e - pid.ePre1

	pid.e = pid.Ke * pid.e
	pid.de = pid.Kde * pid.de

	//将误差e模糊化
	j := 0
	for i := 0; i < N; i++ {
		if pid.mfTE == "trimf" {
			uE[i] = pid.TrimF(pid.e, pid.eMfParas[i*3], pid.eMfParas[i*3+1], pid.eMfParas[i*3+2])
		}
		if uE[i] != 0 {
			uEIndex[j] = i
			j = j + 1
		}
	}
	//富余的空间填0
	for ; j < 3; j++ {
		uEIndex[j] = 0
	}

	//将误差变化率de模糊化
	j = 0
	for i := 0; i < N; i++ {
		if pid.mfTDe == "trimf" {
			uDe[i] = pid.TrimF(pid.de, pid.deMfParas[i*3], pid.deMfParas[i*3+1], pid.deMfParas[i*3+2])
		}
		if uDe[i] != 0 {
			uDeIndex[j] = i
			j = j + 1
		}
	}
	for ; j < 3; j++ {
		uDeIndex[j] = 0
	}

	// 计算delta_Kp和Kp 解模糊
	var den float64 = 0
	var num float64 = 0

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			num += uE[uEIndex[i]] * uDe[uDeIndex[j]] * pid.KpRuleMatrix[uEIndex[i]][uDeIndex[j]]
			den += uE[uEIndex[i]] * uDe[uDeIndex[j]]
		}
	}
	deltaKp = num / den
	deltaKp = pid.KuP * deltaKp

	if deltaKp >= pid.deltaKpMax {
		deltaKp = pid.deltaKpMax
	} else if deltaKp <= -pid.deltaKpMax {
		deltaKp = -pid.deltaKpMax
	}

	if deltaKp == math.NaN() {
		deltaKp = 0
	}

	pid.Kp += deltaKp
	if pid.Kp < 0 {
		pid.Kp = 0
	}

	// 计算delta_Ki和Ki 解模糊
	den = 0
	num = 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			num += uE[uEIndex[i]] * uDe[uDeIndex[j]] * pid.KiRuleMatrix[uEIndex[i]][uDeIndex[j]]
			den += uE[uEIndex[i]] * uDe[uDeIndex[j]]
		}
	}
	deltaKi = num / den
	deltaKi = pid.KuI * deltaKi
	if deltaKi >= pid.deltaKiMax {
		deltaKi = pid.deltaKiMax
	} else if deltaKi <= -pid.deltaKiMax {
		deltaKi = -pid.deltaKiMax
	}
	if deltaKi == math.NaN() {
		deltaKi = 0
	}
	pid.Ki += deltaKi
	if pid.Ki < 0 {
		pid.Ki = 0
	}
	// 计算delta_Kd和Kd 解模糊
	den = 0
	num = 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			num += uE[uEIndex[i]] * uDe[uDeIndex[j]] * pid.KdRuleMatrix[uEIndex[i]][uDeIndex[j]]
			den += uE[uEIndex[i]] * uDe[uDeIndex[j]]
		}
	}
	deltaKd = num / den
	deltaKd = pid.KuD * deltaKd
	if deltaKd >= pid.deltaKdMax {
		deltaKd = pid.deltaKdMax
	} else if deltaKd <= -pid.deltaKdMax {
		deltaKd = -pid.deltaKdMax
	}
	if deltaKd == math.NaN() {
		deltaKd = 0
	}
	pid.Kd += deltaKd
	if pid.Kd < 0 {
		pid.Kd = 0
	}

	//Ki会不断的累计积分，会变得非常大，这里适当缩小Ki的值
	if pid.Ki > 1.2 {
		pid.Ki /= 2
	}

	pid.A = pid.Kp + pid.Ki + pid.Kd
	pid.B = -2*pid.Kd - pid.Kp
	pid.C = pid.Kd

	deltaU = pid.A*pid.e + pid.B*pid.ePre1 + pid.C*pid.ePre2
	deltaU = deltaU / pid.Ke

	if deltaU >= 0.95*pid.target {
		deltaU = 0.95 * pid.target
	} else if deltaU <= -0.95*pid.target {
		deltaU = -0.95 * pid.target
	}

	pid.ePre2 = pid.ePre1
	pid.ePre1 = pid.e

	return deltaU
}
