package manifest

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// MaxNameLength is our longest app name to date. This can be updated, but it will need to
	// be tested in the mobile app.
	MaxNameLength = 32

	// MaxSummaryLength is our longest app summary to date. This can be updated, but it will need to
	// be tested in the mobile app.
	MaxSummaryLength = 32

	dash = '-'
)

// Categories is the list of allowed app categories. Keep it in sync with
// categories.yaml in tronbyt/apps.
var Categories = []string{
	"art",
	"clocks",
	"education",
	"entertainment",
	"finance",
	"food",
	"gaming",
	"hardware",
	"health",
	"lifestyle",
	"news",
	"reference",
	"science",
	"smart-home",
	"social",
	"sports",
	"technology",
	"transit",
	"travel",
	"utilities",
	"weather",
}

var punctuation []string = []string{
	".",
	"!",
	"?",
}

var (
	titleCaser     cases.Caser
	titleCaserOnce sync.Once
)

func getTitleCaser() cases.Caser {
	titleCaserOnce.Do(func() {
		titleCaser = cases.Title(language.English, cases.NoLower)
	})
	return titleCaser
}

// ValidateName ensures the app name provided adheres to the standards for app
// names. We're picky here because these will display in the Tidbyt mobile app
// and need to display properly.
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if name != titleCase(name) {
		return fmt.Errorf("'%s' should be title case, 'Fuzzy Clock' for example", name)
	}

	if len(name) > MaxNameLength {
		return fmt.Errorf("app names need to be less than %d characters", MaxNameLength)
	}

	return nil
}

// ValidateSummary ensures the app summary provided adheres to the standards
// for app summaries. We're picky here because these will display in the Tidbyt
// mobile app and need to display properly.
func ValidateSummary(summary string) error {
	if summary == "" {
		return fmt.Errorf("summary cannot be empty")
	}

	if len(summary) > MaxSummaryLength {
		return fmt.Errorf("app summaries need to be less than %d characters", MaxSummaryLength)
	}

	for _, punct := range punctuation {
		if strings.HasSuffix(summary, punct) {
			return fmt.Errorf("app summaries should not end in punctuation")
		}
	}

	words := strings.Split(summary, " ")
	if len(words) > 0 && words[0] != getTitleCaser().String(words[0]) {
		return fmt.Errorf("app summaries should start with an uppercased character")
	}

	return nil
}

// ValidateDesc ensures the app description provided adheres to the standards
// for app descriptions. We're picky here because these will display in the
// Tidbyt mobile app and need to display properly.
func ValidateDesc(desc string) error {
	if desc == "" {
		return fmt.Errorf("desc cannot be empty")
	}

	found := false
	for _, punct := range punctuation {
		if strings.HasSuffix(desc, punct) {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("app descriptions should end in punctuation")
	}

	words := strings.Split(desc, " ")
	if len(words) > 0 && words[0] != getTitleCaser().String(words[0]) {
		return fmt.Errorf("app descriptions should start with an uppercased character")
	}

	return nil
}

// ValidateAuthor ensures the app author provided adheres to the standards
// for app author. We're picky here because these will display in the
// Tidbyt mobile app and need to display properly.
func ValidateAuthor(author string) error {
	if author == "" {
		return fmt.Errorf("author cannot be empty")
	}

	// I don't know what validation where need here just yet. We're going to
	// have to eyeball it in pull requests until we get a sense of what doesn't
	// work.
	return nil
}

// ValidateCategory ensures the category is one of the allowed categories. It
// is not part of Manifest.Validate because only apps destined for tronbyt/apps
// need one.
func ValidateCategory(category string) error {
	if category == "" {
		return fmt.Errorf("category cannot be empty, choose one of: %s", strings.Join(Categories, ", "))
	}

	if !slices.Contains(Categories, category) {
		return fmt.Errorf("unknown category '%s', choose one of: %s", category, strings.Join(Categories, ", "))
	}

	return nil
}

// ResolveCategory turns a category choice into a category name. The choice may
// be the 1-based number shown by CategoryGrid or the category name itself.
func ResolveCategory(input string) (string, error) {
	input = strings.TrimSpace(input)

	if n, err := strconv.Atoi(input); err == nil {
		if n < 1 || n > len(Categories) {
			return "", fmt.Errorf("enter a number between 1 and %d, or a category name", len(Categories))
		}
		return Categories[n-1], nil
	}

	if err := ValidateCategory(input); err != nil {
		return "", fmt.Errorf("enter a number between 1 and %d, or a category name", len(Categories))
	}
	return input, nil
}

// CategoryGrid renders the categories as numbered columns, filled top to
// bottom, so they can all be shown at once.
func CategoryGrid(columns int) string {
	rows := (len(Categories) + columns - 1) / columns

	widths := make([]int, columns)
	for i, c := range Categories {
		widths[i/rows] = max(widths[i/rows], len(c))
	}

	var b strings.Builder
	for r := range rows {
		for col := range columns {
			i := col*rows + r
			if i >= len(Categories) {
				continue
			}
			cell := fmt.Sprintf("%2d) %s", i+1, Categories[i])
			if col < columns-1 {
				cell = fmt.Sprintf("%-*s", widths[col]+4, cell)
			}
			b.WriteString("  " + cell)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// ValidateID ensures the id will parse when we go to add it to our database
// internally.
func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if id != strings.ToLower(id) {
		return fmt.Errorf("ids should be lower case, %s != %s", id, strings.ToLower(id))
	}

	for _, r := range id {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != dash {
			return fmt.Errorf("ids can only contain letters, numbers, or a dash character")
		}
	}

	return nil
}

func titleCase(input string) string {
	words := strings.Split(input, " ")
	smallwords := " a an on the to of "

	for index, word := range words {
		if strings.Contains(smallwords, " "+word+" ") && word != string(word[0]) {
			words[index] = word
		} else {
			words[index] = getTitleCaser().String(word)
		}
	}

	return strings.Join(words, " ")
}
