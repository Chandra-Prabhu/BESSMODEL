package main

import (
	"errors"
	"fmt"
	"math"

	"github.com/xuri/excelize/v2"
)

func main() {
	fmt.Println("Enter PPA length, tariff")
	var pPALength int = 25
	var constrperiod int = 2
	var tariff float64 = 2.6
	var intrate float64 = 0.085
	var mindebtrepay float64 = 0.01
	var repaymethod string = "Equa"
	var dscr float64 = 1.2
	var payables float64 = 0.12
	var receivables float64 = 0.12
	var debttenure int = 21
	var capacity float64 = 300
	var unitCapex float64 = 4.0
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
	var dsra float64 = 0.075
	generation := constrappend(gencal(capacity, cuf, degradation, pPALength), constrperiod)
	tarifftimeseries := constrappend(tariffcal(tariff, tariffEscalation, pPALength), constrperiod)
	opex := constrappend(tariffcal(unitOpex*capacity*(1.0+gstrate)*costunit, opexEscalation, pPALength), constrperiod)
	revenue := revenuecal(generation, tarifftimeseries)
	ebitda := minus(revenue, opex)
	capex := capacity * unitCapex * costunit
	capexTS := make([]float64, constrperiod)
	capexTS[0] = capex / float64(constrperiod)
	capexTS[1] = capex / float64(constrperiod)
	initialloan := capex * de
	debtrepayment, debtopening, debtoutstanding, interestpayment, dscrts := debtrepay(initialloan, debttenure, repaymethod, ebitda[2:], intrate, dscr, mindebtrepay)
	debtrepayment = constrappend(debtrepayment, constrperiod)
	debtopening = constrappend(debtopening, constrperiod)
	debtoutstanding = constrappend(debtoutstanding, constrperiod)
	interestpayment = constrappend(interestpayment, constrperiod)
	pbdt := minus(ebitda, interestpayment)
	workingcapital, changeinWC := workingcapcal(revenue, opex, payables, receivables)
	_, deprec := depreciationslm(capex-nondeprecap, deprerate, pPALength)
	deprec = constrappend(deprec, constrperiod)
	pbt := minus(pbdt, deprec)
	dsraopening, dsraclosing, dsrachange := dsracal(interestpayment, debtrepayment, dsra)
	debtrepayment[0] = -capexTS[0] * de
	debtrepayment[1] = -capexTS[1] * de
	_, txdeprec := depreciationslm(capex-nondeprecap, txdeprerate, pPALength)
	txdeprec = constrappend(txdeprec, constrperiod)
	taxableincome := minus(pbdt, txdeprec)
	taxes := tax(taxableincome, taxrate)
	profits := minus(pbt, taxes)
	fcfe := minus(minus(ebitda, interestpayment), add(add(capexTS, debtrepayment), minus(taxes, add(changeinWC, dsrachange))))
	fmt.Println(profits[0])
	s, _ := IRR(fcfe)
	fmt.Printf("%f is the EIRR for given assumptions", s)
	wb := excelize.NewFile()
	defer func() {
		if err := wb.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	sheet := "Financials"
	wb.SetSheetName("Sheet1", sheet)
	excelstyling(wb, sheet)
	wb.SetColWidth(sheet, "D", "AF", 18)
	i := 2
	i = adddata(generation, "Generation", wb, sheet, i)
	i = adddata(tarifftimeseries, "Tariff", wb, sheet, i)
	i = adddata(revenue, "Revenue", wb, sheet, i)
	i++
	i = adddata(opex, "Opex", wb, sheet, i)
	i = adddata(ebitda, "Ebitda", wb, sheet, i) + 1
	i = adddata(interestpayment, "Interest paid", wb, sheet, i)
	i = adddata(deprec, "Depreciation", wb, sheet, i)
	i = adddata(pbt, "PBT", wb, sheet, i)
	i = adddata(taxes, "Tax", wb, sheet, i)
	i = adddata(profits, "Profits before Dividend", wb, sheet, i) + 2
	i = adddata(capexTS, "capex phasing", wb, sheet, i)
	i = adddata(debtopening, "debtopening", wb, sheet, i)
	i = adddata(debtoutstanding, "debtoutstanding", wb, sheet, i)
	i = adddata(debtrepayment, "debt repayment", wb, sheet, i)
	i = adddata(dscrts, "dscr ratio", wb, sheet, i) + 1
	i = adddata(txdeprec, "tax depreciation", wb, sheet, i)
	i = adddata(taxableincome, "taxable income", wb, sheet, i) + 1
	i = adddata(fcfe, "fcfe", wb, sheet, i) + 1
	i = adddata(dsraopening, "dsra opening", wb, sheet, i)
	i = adddata(dsraclosing, "dsra closing", wb, sheet, i)
	i = adddata(dsrachange, "dsra change", wb, sheet, i) + 1
	i = adddata(workingcapital, "working capital", wb, sheet, i)
	_ = adddata(changeinWC, "change in WC", wb, sheet, i)
	highlightrows := []int{4, 7, 11, 13, 25}
	highlightstyle(wb, sheet, highlightrows)
	wb.SaveAs("test2.2.xlsx")
}

func celladdress(col int, row int) string {
	var str string
	if col < 26 {
		str = fmt.Sprintf("%s%d", string(rune(64+col)), row)
	} else {
		str = fmt.Sprintf("%s%d", (string(rune(64+col/26)) + string(rune(65+int(math.Mod(float64(col), 26))))), row)
	}
	return str
}

func adddata(data []float64, dataname string, wb *excelize.File, sheet string, i int) int {
	wb.SetCellStr(sheet, celladdress(1, i), dataname)
	wb.SetSheetRow(sheet, celladdress(4, i), &data)
	return i + 1
}

func highlightstyle(wb *excelize.File, sheet string, k []int) {
	style1, _ := wb.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 12,
			Bold: true,
		},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 6},
		},
		NumFmt: 38,
	})
	for _, i := range k {
		wb.SetRowStyle(sheet, i, i, style1)
	}
}

func excelstyling(wb *excelize.File, sheet string) {
	styledef, _ := wb.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "FFFFFF", Style: 2},
			{Type: "right", Color: "FFFFFF", Style: 2},
			{Type: "top", Color: "FFFFFF", Style: 2},
			{Type: "bottom", Color: "FFFFFF", Style: 2},
		},
		NumFmt: 38,
	})
	styletitle, _ := wb.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   12,
			Italic: true,
		},
	})
	wb.SetCellStyle(sheet, "a1", "a100", styletitle)
	wb.SetRowStyle(sheet, 1, 100, styledef)
}

func workingcapcal(revenue []float64, opex []float64, payables float64, receivables float64) ([]float64, []float64) {
	workingcap := make([]float64, len(revenue))
	changeinwc := make([]float64, len(revenue))
	for i := 0; i < len(revenue); i++ {
		workingcap[i] = revenue[i]*payables - opex[i]*receivables
		if i == 0 {
			changeinwc[i] = 0
		} else {
			changeinwc[i] = workingcap[i-1] - workingcap[i]
		}
	}
	return workingcap, changeinwc
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

func debtrepay(initialloan float64, debttenure int, method string, ebitda1 []float64, intrate float64, dscr float64, mindebtrepay float64) ([]float64, []float64, []float64, []float64, []float64) {
	//fmt.Println(ebitda1)
	debtrepayment := make([]float64, debttenure)
	debtoutstanding := make([]float64, debttenure)
	debtopening := make([]float64, debttenure)
	interest := make([]float64, debttenure)
	dscrts := make([]float64, debttenure)
	if method == "Equal" {
		debtopening[0] = initialloan
		debtrepayment[0] = initialloan / float64(debttenure)
		debtoutstanding[0] = debtopening[0] - debtrepayment[0]
		interest[0] = (debtopening[0] + debtoutstanding[0]) / 2.0 * intrate
		dscrts[0] = ebitda1[0] / (debtrepayment[0] + interest[0])
		for i := 1; i < debttenure; i++ {
			debtrepayment[i] = initialloan / float64(debttenure)
			debtopening[i] = debtoutstanding[i-1]
			debtoutstanding[i] = debtopening[i] - debtrepayment[i]
			interest[i] = (debtopening[i] + debtoutstanding[i]) / 2.0 * intrate
			dscrts[i] = ebitda1[i] / (debtrepayment[i] + interest[i])
		}
	} else {
		maxpay := maxrepay(ebitda1[:debttenure], dscr, intrate)
		for i := 0; i < debttenure; i++ {
			if i == 0 {
				debtopening[i] = initialloan
			} else {
				debtopening[i] = debtoutstanding[i-1]
			}
			if (debtopening[i] < maxpay[i]) && ((debtopening[i] - mindebtrepay*initialloan) < maxpay[i+1]) {
				debtrepayment[i] = mindebtrepay * initialloan
			} else if debtopening[i] < maxpay[i] {
				debtrepayment[i] = debtopening[i] - maxpay[i+1]
			} else {
				debtrepayment[i] = (ebitda1[i]/dscr - intrate*debtopening[i]) / (1 - intrate/2.0)
			}
			debtoutstanding[i] = debtopening[i] - debtrepayment[i]
			interest[i] = (debtopening[i] + debtoutstanding[i]) / 2.0 * intrate
			dscrts[i] = ebitda1[i] / (debtrepayment[i] + interest[i])
		}
	}
	return debtrepayment, debtopening, debtoutstanding, interest, dscrts
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

func dsracal(interestpayment []float64, debtrepayment []float64, dsra float64) ([]float64, []float64, []float64) {
	dsraopening := make([]float64, len(interestpayment))
	dsraclosing := make([]float64, len(interestpayment))
	dsrachange := make([]float64, len(interestpayment))
	for i := 0; i < len(interestpayment); i++ {
		if i < 1 {
			dsraopening[i] = 0
		} else {
			dsraopening[i] = dsraclosing[i-1]
		}
		dsraclosing[i] = (interestpayment[i] + debtrepayment[i]) * dsra
		dsrachange[i] = dsraopening[i] - dsraclosing[i]
	}
	return dsraopening, dsraclosing, dsrachange
}
