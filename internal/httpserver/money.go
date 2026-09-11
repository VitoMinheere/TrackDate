package httpserver

import (
	"fmt"
	"strconv"
	"strings"
)

// parseCost accepts plain decimal input like "125", "125.5", or "125.50"
// and returns the amount in integer cents.
func parseCost(input string) (int64, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, nil
	}

	neg := false
	if strings.HasPrefix(input, "-") {
		neg = true
		input = input[1:]
	}

	whole, frac, hasFrac := strings.Cut(input, ".")
	if whole == "" {
		whole = "0"
	}
	wholeCents, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid cost %q", input)
	}

	var fracCents int64
	if hasFrac {
		frac = (frac + "00")[:2]
		fracCents, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid cost %q", input)
		}
	}

	total := wholeCents*100 + fracCents
	if neg {
		total = -total
	}
	return total, nil
}

func formatCost(cents int64, currency string) string {
	return fmt.Sprintf("%.2f %s", float64(cents)/100, currency)
}
