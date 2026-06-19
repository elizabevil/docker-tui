package detail

import (
	"fmt"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
)

func isImageDetailContent(content string) bool {
	return strings.Contains(content, "Tags:") && strings.Contains(content, "Architecture:")
}

func buildImageDetailSections(content string) []detailSection {
	sections := []detailSection{
		{Title: i18n.T("inspect.section_basic"), Lines: []string{}},
		{Title: i18n.T("inspect.section_system"), Lines: []string{}},
		{Title: i18n.T("inspect.section_config"), Lines: []string{}},
		{Title: i18n.T("inspect.section_storage"), Lines: []string{}},
		{Title: i18n.T("inspect.section_tags"), Lines: []string{}},
		{Title: i18n.T("inspect.section_metadata"), Lines: []string{}},
		{Title: i18n.T("inspect.section_history"), Lines: []string{}},
	}

	lines := strings.Split(content, "\n")
	current := "summary"
	kv := make(map[string]string)
	env := make([]string, 0, 16)
	vols := make([]string, 0, 8)
	health := make([]string, 0, 8)
	labels := make([]string, 0, 16)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		switch line {
		case "── System ──":
			current = "system"
			continue
		case "── Runtime ──", "── Entrypoint / Cmd ──":
			current = "config"
			continue
		case "── Volumes ──", "── Healthcheck ──":
			if line == "── Volumes ──" {
				current = "volumes"
			} else {
				current = "health"
			}
			continue
		case "── Storage ──":
			current = "storage"
			continue
		case "── Labels ──":
			current = "labels"
			continue
		}

		if strings.HasPrefix(line, "── Environment (") && strings.HasSuffix(line, " vars) ──") {
			current = "env"
			continue
		}

		if current == "env" {
			env = append(env, line)
			continue
		}
		if current == "volumes" {
			vols = append(vols, line)
			continue
		}
		if current == "health" {
			health = append(health, line)
			continue
		}
		if current == "labels" {
			labels = append(labels, line)
			continue
		}

		if idx := strings.Index(line, ":"); idx > 0 {
			k := strings.TrimSpace(line[:idx])
			v := strings.TrimSpace(line[idx+1:])
			kv[k] = v
		}

		switch {
		case current == "summary":
			if !strings.Contains(line, ":") {
				sections[0].Lines = append(sections[0].Lines, line)
			}
		case current == "system":
			if !strings.Contains(line, ":") {
				sections[1].Lines = append(sections[1].Lines, line)
			}
		case current == "config":
			if !strings.Contains(line, ":") {
				sections[2].Lines = append(sections[2].Lines, line)
			}
		case current == "storage":
			if !strings.Contains(line, ":") {
				sections[3].Lines = append(sections[3].Lines, line)
			}
		default:
			sections[5].Lines = append(sections[5].Lines, line)
		}
	}

	appendKV := func(section *detailSection, key string, label string) {
		if v, ok := kv[key]; ok && v != "" {
			section.Lines = append(section.Lines, fmt.Sprintf("%s: %s", label, v))
		}
	}

	appendKV(&sections[0], "ID", "ID")
	appendKV(&sections[0], "Tags", i18n.T("inspect.tags"))
	appendKV(&sections[0], "Registry", i18n.T("inspect.registry"))
	appendKV(&sections[0], "Name", i18n.T("inspect.name"))
	appendKV(&sections[0], "Tag", i18n.T("inspect.tag"))
	appendKV(&sections[0], "Digest", i18n.T("inspect.digest"))
	appendKV(&sections[0], "Created", i18n.T("inspect.created"))
	appendKV(&sections[0], "Size", i18n.T("inspect.size"))

	appendKV(&sections[1], "Architecture", i18n.T("inspect.arch"))
	appendKV(&sections[1], "OS", "OS")
	appendKV(&sections[1], "OS Version", "OS Version")

	appendKV(&sections[2], "Author", i18n.T("inspect.author"))
	appendKV(&sections[2], "Comment", i18n.T("inspect.comment"))
	appendKV(&sections[2], "WorkingDir", i18n.T("inspect.working_dir"))
	appendKV(&sections[2], "User", i18n.T("inspect.user"))
	appendKV(&sections[2], "StopSignal", i18n.T("inspect.stop_signal"))
	appendKV(&sections[2], "Entrypoint", i18n.T("inspect.entrypoint"))
	appendKV(&sections[2], "Cmd", i18n.T("inspect.cmd"))
	appendKV(&sections[2], "Shell", i18n.T("inspect.shell"))
	appendKV(&sections[2], "OnBuild", i18n.T("inspect.on_build"))
	appendKV(&sections[2], "ExposedPorts", i18n.T("inspect.exposed_ports"))
	if len(env) > 0 {
		sections[2].Lines = append(sections[2].Lines, i18n.T("inspect.environment")+":")
		for _, e := range env {
			sections[2].Lines = append(sections[2].Lines, "  "+e)
		}
	}
	if len(vols) > 0 {
		sections[2].Lines = append(sections[2].Lines, i18n.T("inspect.volumes")+":")
		for _, v := range vols {
			sections[2].Lines = append(sections[2].Lines, "  "+v)
		}
	}
	if len(health) > 0 {
		sections[2].Lines = append(sections[2].Lines, i18n.T("inspect.healthcheck")+":")
		for _, h := range health {
			sections[2].Lines = append(sections[2].Lines, "  "+h)
		}
	}

	appendKV(&sections[3], "Driver", i18n.T("inspect.driver"))
	appendKV(&sections[3], "Layers", i18n.T("inspect.layers"))

	for _, line := range labels {
		sections[4].Lines = append(sections[4].Lines, line)
	}

	sections[5].Lines = append(sections[5].Lines,
		i18n.T("inspect.source"),
		fmt.Sprintf("Fields parsed: %d", len(kv)),
	)

	if len(sections[5].Lines) == 0 {
		sections[5].Lines = []string{i18n.T("inspect.no_metadata")}
	}
	if len(sections[6].Lines) == 0 {
		sections[6].Lines = []string{i18n.T("inspect.no_history")}
	}

	for i := range sections {
		if len(sections[i].Lines) == 0 {
			sections[i].Lines = []string{"—"}
		}
	}

	return sections
}
