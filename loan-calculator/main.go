package main

import (
	"flag"
	"fmt"
	"math"
)

func main() {
	// Parse flags
	payment := flag.Float64("payment", 0, "Monthly payment")
	principal := flag.Float64("principal", 0, "Loan principal")
	periods := flag.Int("periods", 0, "Number of months")
	interest := flag.Float64("interest", 0, "Annual interest rate (without % sign)")
	loanType := flag.String("type", "", "Payment type: annuity or diff")

	flag.Parse()

	// Validation
	if *loanType != "annuity" && *loanType != "diff" {
		fmt.Println("Incorrect parameters")
		return
	}
	if *interest <= 0 {
		fmt.Println("Incorrect parameters")
		return
	}
	if *principal < 0 || *payment < 0 || *periods < 0 {
		fmt.Println("Incorrect parameters")
		return
	}

	// Count provided parameters
	count := 0
	if *payment > 0 {
		count++
	}
	if *principal > 0 {
		count++
	}
	if *periods > 0 {
		count++
	}
	if *interest > 0 {
		count++
	}
	if count < 3 { // type + at least 3 others
		fmt.Println("Incorrect parameters")
		return
	}

	i := *interest / (12 * 100) // monthly interest rate

	// Differentiated payments
	if *loanType == "diff" {
		if *payment > 0 {
			fmt.Println("Incorrect parameters")
			return
		}
		if *principal == 0 || *periods == 0 {
			fmt.Println("Incorrect parameters")
			return
		}
		total := 0
		for m := 1; m <= *periods; m++ {
			d := (*principal / float64(*periods)) +
				i*(*principal-float64(m-1)**principal/float64(*periods))
			pay := int(math.Ceil(d))
			fmt.Printf("Month %d: payment is %d\n", m, pay)
			total += pay
		}
		overpayment := total - int(*principal)
		fmt.Printf("\nOverpayment = %d\n", overpayment)
		return
	}

	// Annuity payments
	if *loanType == "annuity" {
		// Case 1: calculate number of periods
		if *periods == 0 && *principal > 0 && *payment > 0 {
			n := math.Log(*payment/(*payment-i**principal)) / math.Log(1+i)
			months := int(math.Ceil(n))
			years := months / 12
			remMonths := months % 12
			if years > 0 && remMonths > 0 {
				fmt.Printf("It will take %d years and %d months to repay this loan!\n", years, remMonths)
			} else if years > 0 {
				fmt.Printf("It will take %d years to repay this loan!\n", years)
			} else {
				fmt.Printf("It will take %d months to repay this loan!\n", remMonths)
			}
			overpayment := int(*payment)*months - int(*principal)
			fmt.Printf("Overpayment = %d\n", overpayment)
			return
		}

		// Case 2: calculate monthly payment
		if *payment == 0 && *principal > 0 && *periods > 0 {
			n := float64(*periods)
			A := *principal * (i * math.Pow(1+i, n)) / (math.Pow(1+i, n) - 1)
			pay := int(math.Ceil(A))
			fmt.Printf("Your annuity payment = %d!\n", pay)
			overpayment := pay*int(n) - int(*principal)
			fmt.Printf("Overpayment = %d\n", overpayment)
			return
		}

		// Case 3: calculate principal
		if *principal == 0 && *payment > 0 && *periods > 0 {
			n := float64(*periods)
			P := *payment / ((i * math.Pow(1+i, n)) / (math.Pow(1+i, n) - 1))
			pr := int(math.Floor(P))
			fmt.Printf("Your loan principal = %d!\n", pr)
			overpayment := int(*payment)*int(n) - pr
			fmt.Printf("Overpayment = %d\n", overpayment)
			return
		}
	}

	// If none matched
	fmt.Println("Incorrect parameters")
}
