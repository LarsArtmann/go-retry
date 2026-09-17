package retry_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode"
)

// The docs gates below exist because the 2026-09-17 rendering catastrophe
// shipped three strikethrough bug classes and two unannotated archive items
// with no gate reading the documentation. Like TestModuleGoDirectiveStaysPinned,
// they turn documented conventions into enforced invariants: the rendering
// rules live in docs/status/README.md, the annotation convention next to them.

type docIssue struct {
	kind string
	file string
	line int
	ctx  string
}

func repoMarkdownFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".git" || strings.HasPrefix(d.Name(), ".") {
				if path != "." {
					return filepath.SkipDir
				}

				return nil
			}

			return nil
		}
		if strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walking repo for markdown files: %v", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		t.Fatalf("no markdown files found; the walk is broken")
	}

	return files
}

// stripInlineCode removes paired code spans so their literal ~~ and ` never
// trip the scanners; an unpaired run is preserved so the cell-balance check
// can see it.
func stripInlineCode(line string) string {
	var out strings.Builder
	i := 0
	for i < len(line) {
		if line[i] != '`' {
			out.WriteByte(line[i])
			i++

			continue
		}
		runEnd := i
		for runEnd < len(line) && line[runEnd] == '`' {
			runEnd++
		}
		run := line[i:runEnd]
		if closeIdx := strings.Index(line[runEnd:], run); closeIdx >= 0 {
			out.WriteByte(0)
			i = runEnd + closeIdx + len(run)
		} else {
			out.WriteString(run)
			i = runEnd
		}
	}

	return out.String()
}

func splitTableCells(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	var cells []string
	var cell strings.Builder
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] == '\\' && i+1 < len(trimmed) && trimmed[i+1] == '|' {
			cell.WriteString(`\|`)
			i++

			continue
		}
		if trimmed[i] == '|' {
			cells = append(cells, cell.String())
			cell.Reset()

			continue
		}
		cell.WriteByte(trimmed[i])
	}

	return append(cells, cell.String())
}

var (
	docFenceLine  = regexp.MustCompile("^\\s*(`{3,}|~{3,})")
	docHeaderLine = regexp.MustCompile(`^#{1,6}\s`)
	docHRLine     = regexp.MustCompile(`^\s*(-{3,}|\*{3,}|_{3,})\s*$`)
)

func isTableLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "|")
}

func scanSpanText(text string, file string, baseLine int, issues *[]docIssue) {
	lineOffset := func(pos int) int {
		return baseLine + strings.Count(text[:pos], "\n")
	}
	openPos := -1
	for i := 0; i < len(text); {
		if text[i] != '~' {
			i++

			continue
		}
		if i+1 < len(text) && text[i+1] == '~' {
			if openPos < 0 {
				openPos = i
			} else {
				if i > 0 {
					switch text[i-1] {
					case ' ', '\t', '\n':
						*issues = append(*issues, docIssue{
							kind: "space-before-closer (strike renders literally; glue the closer to the last struck character)",
							file: file,
							line: lineOffset(openPos),
							ctx:  clip(text[max(0, i-40) : i+12]),
						})
					}
				}
				openPos = -1
			}
			i += 2

			continue
		}
		if openPos >= 0 {
			*issues = append(*issues, docIssue{
				kind: "lone-tilde-in-span (kills the pairing; write ≈ instead)",
				file: file,
				line: lineOffset(i),
				ctx:  clip(text[max(0, i-40) : i+12]),
			})
		}
		i++
	}
	if openPos >= 0 {
		*issues = append(*issues, docIssue{
			kind: "unclosed-span (swallows following text into the strike)",
			file: file,
			line: lineOffset(openPos),
			ctx:  clip(text[openPos:min(len(text), openPos+60)]),
		})
	}
}

func clip(raw string) string {
	s := strings.ReplaceAll(raw, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 70 {
		return s[:70] + "…"
	}

	return s
}

func scanMarkdownFile(path string, lines []string) []docIssue {
	var issues []docIssue
	i := 0
	for i < len(lines) {
		line := lines[i]
		if match := docFenceLine.FindStringSubmatch(line); match != nil {
			fence := match[1][:3]
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), fence) {
				i++
			}
			i++

			continue
		}
		if isTableLine(line) {
			for _, cell := range splitTableCells(line) {
				clean := stripInlineCode(cell)
				if strings.ContainsRune(clean, '`') {
					issues = append(issues, docIssue{
						kind: "unclosed-code-span-in-cell (backticks pair per table cell)",
						file: path,
						line: i + 1,
						ctx:  clip(cell),
					})
				}
				scanSpanText(clean, path, i+1, &issues)
			}
			i++

			continue
		}
		if strings.TrimSpace(line) == "" || docHeaderLine.MatchString(strings.TrimLeft(line, " ")) ||
			docHRLine.MatchString(line) {
			i++

			continue
		}
		paraStart := i
		var para []string
		for i < len(lines) && strings.TrimSpace(lines[i]) != "" &&
			!docFenceLine.MatchString(lines[i]) && !isTableLine(lines[i]) &&
			!docHeaderLine.MatchString(strings.TrimLeft(lines[i], " ")) {
			para = append(para, stripInlineCode(lines[i]))
			i++
		}
		scanSpanText(strings.Join(para, "\n"), path, paraStart+1, &issues)
	}

	return issues
}

func TestMarkdownStrikethroughSpansRender(t *testing.T) {
	t.Parallel()
	var issues []docIssue
	for _, path := range repoMarkdownFiles(t) {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		issues = append(issues, scanMarkdownFile(path, strings.Split(string(raw), "\n"))...)
	}
	for _, issue := range issues {
		t.Errorf("%s:%d: %s: %q", issue.file, issue.line, issue.kind, issue.ctx)
	}
	if len(issues) == 0 {
		t.Log("all markdown strikethrough spans are renderable")
	}
}

func TestMarkdownTableCellsCloseCodeSpans(t *testing.T) {
	t.Parallel()
	var issues []docIssue
	for _, path := range repoMarkdownFiles(t) {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		for idx, line := range strings.Split(string(raw), "\n") {
			if !isTableLine(line) {
				continue
			}
			for _, cell := range splitTableCells(line) {
				if clean := stripInlineCode(cell); strings.ContainsRune(clean, '`') {
					issues = append(issues, docIssue{
						kind: "unclosed-code-span-in-cell",
						file: path,
						line: idx + 1,
						ctx:  clip(cell),
					})
				}
			}
		}
	}
	for _, issue := range issues {
		t.Errorf("%s:%d: %s: %q", issue.file, issue.line, issue.kind, issue.ctx)
	}
}

var (
	archivedSection = regexp.MustCompile(`^##\s+([a-z])\)`)
	numberedItem    = regexp.MustCompile(`^\s{0,3}\d+\.\s`)
	tableSeparator  = regexp.MustCompile(`^:?-{2,}:?$`)
)

func isVerdictSection(letter string) bool {
	return letter == "b" || letter == "c" || letter == "f" || letter == "g"
}

type archivedItem struct {
	file    string
	line    int
	section string
	text    string
}

// collectNumberedItems segments §b/§c/§f/§g numbered list items. An item ends
// at the next item, header, or table; a blank line ends it unless the next
// line is indented (loose-list continuation) or another list element.
func collectNumberedItems(path string, lines []string) []archivedItem {
	var items []archivedItem
	section := ""
	var item *archivedItem
	flush := func() {
		if item != nil && isVerdictSection(section) {
			items = append(items, *item)
		}
		item = nil
	}
	for idx := range len(lines) {
		line := lines[idx]
		if match := archivedSection.FindStringSubmatch(line); match != nil {
			flush()
			section = match[1]

			continue
		}
		if numberedItem.MatchString(line) {
			flush()
			item = &archivedItem{file: path, line: idx + 1, section: section, text: strings.TrimSpace(line)}

			continue
		}
		if item == nil {
			continue
		}
		switch {
		case strings.TrimSpace(line) == "":
			next := ""
			if idx+1 < len(lines) {
				next = lines[idx+1]
			}
			if strings.HasPrefix(next, " ") || strings.HasPrefix(next, "\t") ||
				numberedItem.MatchString(next) || isTableLine(next) {
				item.text += " " + strings.TrimSpace(line)
			} else {
				flush()
			}
		case isTableLine(line) || archivedSection.MatchString(line) || docHeaderLine.MatchString(strings.TrimLeft(line, " ")):
			flush()
		default:
			item.text += " " + strings.TrimSpace(line)
		}
	}
	flush()

	return items
}

var verdictPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)done at`),
	regexp.MustCompile(`(?i)routed to`),
	regexp.MustCompile(`(?i)won['’]t[-\s](?:implement|for\s+now)`),
	regexp.MustCompile(`(?i)not-do`),
	regexp.MustCompile(`(?i)duplicate`),
	regexp.MustCompile(`(?i)\bverified\b`),
	regexp.MustCompile(`(?i)\bresolved\b`),
	regexp.MustCompile(`(?i)\bsuperseded\b`),
	regexp.MustCompile(`(?i)\bdeferred\b`),
	regexp.MustCompile(`(?i)\bdecided\b`),
	regexp.MustCompile(`(?i)\bno-op\b`),
	regexp.MustCompile(`(?i)covered by`),
	regexp.MustCompile(`(?i)nothing to do`),
	regexp.MustCompile(`(?i)correct per convention`),
	regexp.MustCompile(`[.)!?:;*(\x00]\s*→`),
	regexp.MustCompile(`^→`),
	regexp.MustCompile(`(?i)[—–]\s*done\b`),
	regexp.MustCompile(`(?i)[.)]\s+open\s*[—–-]`),
}

func itemCarriesVerdict(text string) bool {
	clean := stripInlineCode(text)
	if strings.Contains(clean, "~~") {
		return true
	}
	normalized := strings.Join(strings.FieldsFunc(clean, unicode.IsSpace), " ")
	for _, pattern := range verdictPatterns {
		if pattern.MatchString(normalized) {
			return true
		}
	}

	return false
}

// checkVerdictTables enforces the column-wise verdict convention: in a gated
// section, every row of a table with a Verdict or Status column must fill it.
func checkVerdictTables(path string, lines []string, section string, issues *[]docIssue) {
	i := 0
	for i < len(lines) {
		if !isTableLine(lines[i]) {
			i++
			continue
		}
		start := i
		var block []string
		for i < len(lines) && isTableLine(lines[i]) {
			block = append(block, lines[i])
			i++
		}
		if !isVerdictSection(section) || len(block) < 3 {
			continue
		}
		header := splitTableCells(block[0])
		verdictIdx := -1
		for cellIdx, cell := range header {
			name := strings.ToLower(strings.TrimSpace(cell))
			if name == "verdict" || name == "status" {
				verdictIdx = cellIdx

				break
			}
		}
		if verdictIdx < 0 {
			continue
		}
		for rowIdx, row := range block[2:] {
			cells := splitTableCells(row)
			if verdictIdx >= len(cells) {
				*issues = append(*issues, docIssue{
					kind: "table-row-missing-verdict-cell",
					file: path,
					line: start + rowIdx + 3,
					ctx:  clip(row),
				})
				continue
			}
			if strings.TrimSpace(stripInlineCode(cells[verdictIdx])) == "" {
				*issues = append(*issues, docIssue{
					kind: "empty-verdict-cell",
					file: path,
					line: start + rowIdx + 3,
					ctx:  clip(row),
				})
			}
		}
	}
}

func TestArchivedReportItemsCarryVerdicts(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir("docs/status/archived")
	if err != nil {
		t.Fatalf("reading archive dir: %v", err)
	}
	var issues []docIssue
	itemCount := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".md") {
			// The one archived .html report is out of scope by design: it
			// carries no strikethrough or list markup to gate (zero <del>
			// spans; its Status column is HTML), and the README index row
			// records it as parse-validated instead.
			continue
		}
		path := filepath.Join("docs", "status", "archived", name)
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		lines := strings.Split(string(raw), "\n")
		items := collectNumberedItems(path, lines)
		itemCount += len(items)
		for _, item := range items {
			if !itemCarriesVerdict(item.text) {
				issues = append(issues, docIssue{
					kind: "unannotated-item (add done at <hash> / → routed to / **Won't implement — reason**)",
					file: item.file,
					line: item.line,
					ctx:  clip(item.text),
				})
			}
		}
		section := ""
		sectionStart := 0
		flushTables := func(end int) {
			checkVerdictTables(path, lines[sectionStart:end], section, &issues)
		}
		for idx, line := range lines {
			if match := archivedSection.FindStringSubmatch(line); match != nil {
				flushTables(idx)
				section = match[1]
				sectionStart = idx
			}
		}
		flushTables(len(lines))
	}
	for _, issue := range issues {
		t.Errorf("%s:%d: %s: %q", issue.file, issue.line, issue.kind, issue.ctx)
	}
	t.Logf("gated items checked: %d", itemCount)
}

func TestStatusIndexCoversArchive(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("docs", "status", "README.md"))
	if err != nil {
		t.Fatalf("reading docs/status/README.md: %v", err)
	}
	indexLink := regexp.MustCompile(`\]\((archived/[^)]+)\)`)
	indexed := map[string]bool{}
	for lineIdx, line := range strings.Split(string(raw), "\n") {
		match := indexLink.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		target := match[1]
		indexed[strings.TrimPrefix(target, "archived/")] = true
		if _, statErr := os.Stat(filepath.Join("docs", "status", filepath.FromSlash(target))); statErr != nil {
			t.Errorf("docs/status/README.md:%d: index links to missing file %s", lineIdx+1, target)
		}
		cells := splitTableCells(line)
		if len(cells) >= 3 && strings.TrimSpace(stripInlineCode(cells[2])) == "" {
			t.Errorf("docs/status/README.md:%d: index row has an empty State cell", lineIdx+1)
		}
	}
	entries, err := os.ReadDir(filepath.Join("docs", "status", "archived"))
	if err != nil {
		t.Fatalf("reading archive dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !indexed[entry.Name()] {
			t.Errorf("archived report %s is not listed in docs/status/README.md", entry.Name())
		}
	}
	if len(indexed) == 0 {
		t.Fatalf("index parse found no rows; the README table shape changed")
	}
}
