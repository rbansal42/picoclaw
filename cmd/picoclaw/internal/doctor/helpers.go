package doctor

import (
	"fmt"

	"github.com/sipeed/picoclaw/cmd/picoclaw/internal"
	"github.com/sipeed/picoclaw/pkg/doctor"
)

func runDoctor(fix bool) error {
	fmt.Printf("%s picoclaw doctor\n\n", internal.Logo)

	opts := doctor.Options{
		Fix: fix,
	}

	findings := doctor.Run(opts)

	// Group findings by check
	errors := 0
	warns := 0
	fixed := 0

	for _, f := range findings {
		icon := f.Severity.Icon()
		switch f.Severity {
		case doctor.SeverityInfo:
			fmt.Printf("  [%s] %s\n", icon, f.Message)
		case doctor.SeverityWarn:
			fmt.Printf("  [%s] %s\n", icon, f.Message)
			warns++
		case doctor.SeverityError:
			fmt.Printf("  [%s] %s\n", icon, f.Message)
			errors++
		}

		// Auto-fix if requested and available
		if fix && f.FixFunc != nil {
			fmt.Printf("      -> fixing: %s ... ", f.Fix)
			if err := f.FixFunc(); err != nil {
				fmt.Printf("FAILED: %v\n", err)
			} else {
				fmt.Printf("OK\n")
				fixed++
				// Downgrade the counts since we fixed it
				if f.Severity == doctor.SeverityError {
					errors--
				} else if f.Severity == doctor.SeverityWarn {
					warns--
				}
			}
		}
	}

	fmt.Println()
	if errors == 0 && warns == 0 {
		fmt.Printf("%s All checks passed!\n", internal.Logo)
	} else {
		summary := fmt.Sprintf("%s Found", internal.Logo)
		if errors > 0 {
			summary += fmt.Sprintf(" %d error(s)", errors)
		}
		if warns > 0 {
			if errors > 0 {
				summary += " and"
			}
			summary += fmt.Sprintf(" %d warning(s)", warns)
		}
		if fixed > 0 {
			summary += fmt.Sprintf(" (%d fixed)", fixed)
		}
		fmt.Println(summary)

		// Hint about --fix if there were fixable problems and --fix wasn't used
		if !fix {
			hasFixable := false
			for _, f := range findings {
				if f.FixFunc != nil {
					hasFixable = true
					break
				}
			}
			if hasFixable {
				fmt.Println("  Run 'picoclaw doctor --fix' to attempt automatic fixes")
			}
		}
	}

	if errors > 0 {
		return fmt.Errorf("doctor found %d error(s)", errors)
	}

	return nil
}
