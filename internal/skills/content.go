package skills

import (
	"fmt"
	"strings"
)

// skillMDContent rewrites a pack's embedded markdown into the SKILL.md format
// required by the open Agent Skills standard (https://agentskills.io):
// frontmatter must contain `name` and `description` keys, where name matches
// the parent directory name (here "vip-<pack-id>"). The pack's own
// title/description frontmatter is replaced; the markdown body is preserved.
func skillMDContent(pack Pack, raw []byte) []byte {
	body := stripFrontmatter(string(raw))
	frontmatter := fmt.Sprintf("---\nname: vip-%s\ndescription: %s\n---\n\n", pack.ID, pack.Description)
	return []byte(frontmatter + body)
}

// stripFrontmatter removes a leading "---\n...\n---\n" YAML frontmatter block
// from content, if present, returning the remaining body with leading blank
// lines trimmed.
func stripFrontmatter(content string) string {
	const marker = "---"
	if !strings.HasPrefix(content, marker+"\n") {
		return content
	}

	rest := content[len(marker)+1:]
	closingIdx := strings.Index(rest, "\n"+marker)
	if closingIdx == -1 {
		return content
	}

	afterClosing := rest[closingIdx+len("\n"+marker):]
	if nl := strings.IndexByte(afterClosing, '\n'); nl != -1 {
		return strings.TrimLeft(afterClosing[nl+1:], "\n")
	}
	return strings.TrimLeft(afterClosing, "\n")
}
