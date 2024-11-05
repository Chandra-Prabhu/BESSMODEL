package main

import (
	"errors"
	"fmt"
	"math"

	"github.com/tealeg/xlsx"
)

func main() {
	fmt.Println("Enter PPA length, tariff")
	var pPALength int = 25
	var constrperiod int = 2
	var tariff float64 = 2.6
	var intrate float64 = 0.09
	var mindebtrepay float64 = 0.01
	var repaymethod string = "Equa"
	var dscr float64 = 1.2
	var debttenure int = 21
	var capacity float64 = 300
	var unitCapex float64 = 4.0
	//var irr float64 = 0.18
	var unitOpex float64 = 0.06
	var costunit float64 = 10000000
	var cuf float64 = 0.292
	var degradation float64 = 0.005
	var tariffEscalation float64 = 0.00
	var opexEscalation float64 = 0.04
	var gstrate float64 = 0.18
	var taxrate float64 = 0.265
	var de float64 = 0.75
	var deprerate float64 = 0.06
	var txdeprerate float64 = 0.075
	var nondeprecap float64 = 0
	//fmt.Scanf("%d", &pPALength)
	//fmt.Scanf("%f", &tariff)
	generation := constrappend(gencal(capacity, cuf, degradation, pPALength), constrperiod)
	tarifftimeseries := constrappend(tariffcal(tariff, tariffEscalation, pPALength), constrperiod)
	opex := constrappend(tariffcal(unitOpex*capacity*(1.0+gstrate)*costunit, opexEscalation, pPALength), constrperiod)
	revenue := revenuecal(generation, tarifftimeseries)
	//fmt.Println(generation, "\n1\n", revenue, "\n2\n", opex)
	ebitda := minus(revenue, opex)
	//fmt.Println(revenue[3], ebitda[3])
	capex := capacity * unitCapex * costunit
	capexTS := make([]float64, constrperiod)
	capexTS[0] = capex / float64(constrperiod)
	capexTS[1] = capex / float64(constrperiod)
	initialloan := capex * de
	//fmt.Println(capexTS, initialloan)
	debtrepayment, debtopening, debtoutstanding, interestpayment := debtrepay(initialloan, debttenure, repaymethod, ebitda[2:], intrate, dscr, mindebtrepay)
	debtrepayment = constrappend(debtrepayment, constrperiod)
	debtopening = constrappend(debtopening, constrperiod)
	debtoutstanding = constrappend(debtoutstanding, constrperiod)
	//fmt.Println(debtopening[0], debtoutstanding[0])
	//fmt.Println(debtrepayment, "\n\n", debtopening, "\n\n", interestpayment, "\n\n", debtoutstanding[0])
	interestpayment = constrappend(interestpayment, constrperiod)
	pbdt := minus(ebitda, interestpayment)
	_, deprec := depreciationslm(capex-nondeprecap, deprerate, pPALength)
	//fadp = constrappend(fadp, constrperiod)
	deprec = constrappend(deprec, constrperiod)
	pbt := minus(pbdt, deprec)
	//fmt.Println(interestpayment)
	debtrepayment[0] = -capexTS[0] * de
	debtrepayment[1] = -capexTS[1] * de
	_, txdeprec := depreciationslm(capex-nondeprecap, txdeprerate, pPALength)
	//txfadp = constrappend(txfadp, constrperiod)
	txdeprec = constrappend(txdeprec, constrperiod)
	taxableincome := minus(pbdt, txdeprec)
	taxes := tax(taxableincome, taxrate)
	profits := minus(pbt, taxes)
	//fmt.Println(fadp, "\n\n", deprec, "\n\n", txdeprec)
	fcfe := minus(minus(ebitda, interestpayment), add(add(capexTS, debtrepayment), taxes))
	//print(profits)
	fmt.Println()
	//print(fcfe)
	//fmt.Println(ebitda, "\n\n", interestpayment, "\n\n", capexTS, "\n\n", debtrepayment, "\n\n", fcfe)
	s, _ := IRR(fcfe)
	fmt.Printf("%f is the EIRR for given assumptions", s)
	wb := xlsx.NewFile()
	sh, _ := wb.AddSheet("Financials")
	adddata(generation, "Generation", sh)
	adddata(tarifftimeseries, "Tariff", sh)
	adddata(revenue, "Revenue", sh)
	adddata(opex, "Opex", sh)
	sh.AddRow()
	adddata(ebitda, "Ebitda", sh)
	adddata(interestpayment, "Interest paid", sh)
	adddata(deprec, "Depreciation", sh)
	adddata(pbt, "PBT", sh)
	sh.AddRow()
	adddata(taxes, "Tax", sh)
	sh.AddRow()
	adddata(profits, "Profits before Dividend", sh)
	sh.AddRow()
	sh.AddRow()
	adddata(capexTS, "capex phasing", sh)
	sh.AddRow()
	adddata(debtopening, "debtopening", sh)
	adddata(debtoutstanding, "debtoutstanding", sh)
	adddata(debtrepayment, "debt repayment", sh)
	sh.AddRow()
	adddata(txdeprec, "tax depreciation", sh)
	adddata(taxableincome, "taxable income", sh)
	sh.AddRow()
	adddata(fcfe, "fcfe", sh)
	wb.Save("test2.1.xlsx")
}

func adddata(data []float64, dataname string, sh *xlsx.Sheet) {
	i := sh.AddRow()
	k := i.AddCell()
	k.Value = dataname

	for j := 1; j <= len(data); j++ {
		k := i.AddCell()
		k.SetFloat(data[j-1])
	}
}

func revenuecal(gen []float64, tariff []float64) []float64 {
	revenue := make([]float64, len(gen))
	for i := 0; i < len(gen); i++ {
		revenue[i] = gen[i] * tariff[i]
	}
	return revenue
}

func gencal(cap float64, cuf float64, degrad float64, ppaLength int) []float64 {
	gen := make([]float64, ppaLength)
	for i := 0; i < ppaLength; i++ {
		gen[i] = cap * cuf * 8760.0 * 1000 * math.Pow(1.0-degrad, float64(i))
	}
	return gen
}

func tariffcal(tariff float64, escalation float64, ppaLength int) []float64 {
	tariffts := make([]float64, ppaLength)
	for i := 0; i < ppaLength; i++ {
		tariffts[i] = tariff * math.Pow((1.0+escalation), float64(i))
	}
	return tariffts
}

func minus(from []float64, sub []float64) []float64 {
	size := min(len(from), len(sub))
	to := make([]float64, 0)
	for i := 0; i < size; i++ {
		to = append(to, (from[i] - sub[i]))
	}
	if len(from) < len(sub) {
		for i := size; i < len(sub); i++ {
			to = append(to, -sub[i])
		}
	} else {
		for i := size; i < len(from); i++ {
			to = append(to, from[i])
		}
	}
	return to
}

func add(from []float64, sub []float64) []float64 {
	size := min(len(from), len(sub))
	to := make([]float64, 0)
	for i := 0; i < size; i++ {
		to = append(to, from[i]+sub[i])
	}
	if len(from) < len(sub) {
		for i := size; i < len(sub); i++ {
			to = append(to, sub[i])
		}
	} else {
		for i := size; i < len(from); i++ {
			to = append(to, from[i])
		}
	}
	return to
}

/*func print(data []float64) {
	for i := 0; i < len(data); i++ {
		fmt.Println(data[i])
	}
}*/

/*func interest(debtopening []float64, debtoutstanding []float64, intrate float64) []float64 {
	interest := make([]float64, 0)
	for i := 0; i < len(debtopening); i++ {
		interest = append(interest, (debtopening[i]-debtoutstanding[i])*intrate)
	}
	return interest
}*/

func depreciationslm(deprecapex float64, deprerate float64, pPALength int) ([]float64, []float64) {
	fadp := make([]float64, pPALength)
	deprec := make([]float64, pPALength)
	//var i float64
	fadp[0] = deprecapex - deprecapex*deprerate
	deprec[0] = deprecapex * deprerate
	for j := 1; j < pPALength; j++ {
		if fadp[j-1] > deprecapex*deprerate {
			fadp[j] = fadp[j-1] - deprecapex*deprerate
		} else {
			fadp[j] = 0
		}
		deprec[j] = fadp[j-1] - fadp[j]
	}
	return fadp, deprec
}

func tax(taxableincome []float64, taxrate float64) []float64 {
	taxes := make([]float64, len(taxableincome))
	losscarry := make([]float64, len(taxableincome))
	for i := 0; i < len(taxableincome); i++ {
		if i == 0 {
			if taxableincome[i] < 0 {
				losscarry[i] = taxableincome[i]
			} else {
				taxes[i] = taxableincome[i] * taxrate
			}
		} else if taxableincome[i] < 0 {
			losscarry[i] = losscarry[i-1] + taxableincome[i]
			taxes[i] = 0
		} else if taxableincome[i] < losscarry[i-1] {
			losscarry[i] = losscarry[i-1] - taxableincome[i]
			taxes[i] = 0
		} else {
			losscarry[i] = 0
			taxes[i] = (taxableincome[i] - losscarry[i-1]) * taxrate
		}
	}
	return taxes
}

func constrappend(series []float64, constrperiod int) []float64 {
	toseries := make([]float64, constrperiod+len(series))
	for i := 0; i < (constrperiod + len(series)); i++ {
		if i < constrperiod {
			toseries[i] = 0
		} else {
			toseries[i] = series[i-constrperiod]
		}
	}
	return toseries
}

const (
	irrMaxInterations = 20
	irrAccuracy       = 1e-7
	irrInitialGuess   = 0
)

// IRR returns the Internal Rate of Return (IRR).
func IRR(values []float64) (float64, error) {
	if len(values) == 0 {
		return 0, errors.New("values must include the initial investment (usually negative number) and period cash flows")
	}
	x0 := float64(irrInitialGuess)
	var x1 float64
	for i := 0; i < irrMaxInterations; i++ {
		fValue := float64(0)
		fDerivative := float64(0)
		for k := 0; k < len(values); k++ {
			fk := float64(k)
			fValue += values[k] / math.Pow(1.0+x0, fk)
			fDerivative += -fk * values[k] / math.Pow(1.0+x0, fk+1.0)
		}
		x1 = x0 - fValue/fDerivative
		if math.Abs(x1-x0) <= irrAccuracy {
			return x1, nil
		}
		x0 = x1
	}
	return x0, errors.New("could not find irr for the provided values")
}

func debtrepay(initialloan float64, debttenure int, method string, ebitda1 []float64, intrate float64, dscr float64, mindebtrepay float64) ([]float64, []float64, []float64, []float64) {
	//fmt.Println(ebitda1)
	debtrepayment := make([]float64, debttenure)
	debtoutstanding := make([]float64, debttenure)
	debtopening := make([]float64, debttenure)
	interest := make([]float64, debttenure)
	if method == "Equal" {
		debtopening[0] = initialloan
		debtrepayment[0] = initialloan / float64(debttenure)
		debtoutstanding[0] = debtopening[0] - debtrepayment[0]
		interest[0] = (debtopening[0] + debtoutstanding[0]) / 2.0 * intrate
		for i := 1; i < debttenure; i++ {
			debtrepayment[i] = initialloan / float64(debttenure)
			debtopening[i] = debtoutstanding[i-1]
			debtoutstanding[i] = debtopening[i] - debtrepayment[i]
			interest[i] = (debtopening[i] + debtoutstanding[i]) / 2.0 * intrate
		}
	} else {
		maxpay := maxrepay(ebitda1[:debttenure], dscr, intrate)
		for i := 0; i < debttenure; i++ {
			if i == 0 {
				debtopening[i] = initialloan
			} else {
				debtopening[i] = debtoutstanding[i-1]
			}
			//edbitda2 := ebitda1[i+1 : debttenure]
			if (debtopening[i] < maxpay[i]) && ((debtopening[i] - mindebtrepay*initialloan) < maxpay[i+1]) {
				//fmt.Println(i, "the debt opening is ", debtopening[i], "and maximum debt that future ebitda can sustain is", maxpay)
				debtrepayment[i] = mindebtrepay * initialloan
			} else if debtopening[i] < maxpay[i] {
				debtrepayment[i] = debtopening[i] - maxpay[i+1]
				//debtoutstanding[i] = debtopening[i] - debtrepayment[i]
			} else {
				debtrepayment[i] = (ebitda1[i]/dscr - intrate*debtopening[i]) / (1 - intrate/2.0)
				/*debtoutstanding[i] = debtopening[i] - debtrepayment[i]
				if debtoutstanding[i] < 0 {
					debtrepayment[i] = debtopening[i]
					debtoutstanding[i] = 0
				}*/
			}
			debtoutstanding[i] = debtopening[i] - debtrepayment[i]
			interest[i] = (debtopening[i] + debtoutstanding[i]) / 2.0 * intrate
		}
	}
	return debtrepayment, debtopening, debtoutstanding, interest
}

func maxrepay(ebitda []float64, dscr float64, intrate float64) []float64 {
	maxpay := make([]float64, len(ebitda))
	maxpay[len(ebitda)-1] = 0.0
	for i := len(ebitda) - 2; i >= 0; i-- {
		maxpay[i] = (2*ebitda[i]/dscr + maxpay[i+1]*(2.0-intrate)) / (2.0 + intrate)
	}
	//fmt.Println(maxdebtopening)
	return maxpay
}
