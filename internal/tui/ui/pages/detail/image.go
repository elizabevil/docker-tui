package detail

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/utils"
)

func buildImageDetailDataSections(data *runtimeapi.ImageDetail) []detailSection {
	if data == nil {
		return nil
	}
	sections := make([]detailSection, 0, 7)
	appendValue := func(lines *[]string, label, value string) {
		if value != "" {
			*lines = append(*lines, label+": "+value)
		}
	}

	basic := detailSection{Title: i18n.T("inspect.section_basic")}
	appendValue(&basic.Lines, "ID", data.ID)
	appendValue(&basic.Lines, i18n.T("inspect.tags"), strings.Join(data.RepoTags, ", "))
	appendValue(&basic.Lines, i18n.T("inspect.registry"), data.Registry)
	appendValue(&basic.Lines, i18n.T("inspect.name"), data.Name)
	appendValue(&basic.Lines, i18n.T("inspect.tag"), data.Tag)
	appendValue(&basic.Lines, i18n.T("inspect.digest"), strings.Join(data.RepoDigests, ", "))
	appendValue(&basic.Lines, i18n.T("inspect.created"), data.Created)
	if data.Size > 0 {
		appendValue(&basic.Lines, i18n.T("inspect.size"), utils.FormatBytes(float64(data.Size)))
	}
	sections = append(sections, basic)

	system := detailSection{Title: i18n.T("inspect.section_system")}
	appendValue(&system.Lines, i18n.T("inspect.arch"), data.Architecture)
	appendValue(&system.Lines, "OS", data.OS)
	appendValue(&system.Lines, "OS Version", data.OSVersion)
	appendValue(&system.Lines, i18n.T("inspect.author"), data.Author)
	appendValue(&system.Lines, i18n.T("inspect.comment"), data.Comment)
	if len(system.Lines) > 0 {
		sections = append(sections, system)
	}

	runtime := detailSection{Title: i18n.T("inspect.section_config")}
	appendValue(&runtime.Lines, i18n.T("inspect.working_dir"), data.Runtime.WorkingDir)
	appendValue(&runtime.Lines, i18n.T("inspect.user"), data.Runtime.User)
	appendValue(&runtime.Lines, i18n.T("inspect.stop_signal"), data.Runtime.StopSignal)
	appendValue(&runtime.Lines, i18n.T("inspect.entrypoint"), strings.Join(data.Runtime.Entrypoint, " "))
	appendValue(&runtime.Lines, i18n.T("inspect.cmd"), strings.Join(data.Runtime.Cmd, " "))
	appendValue(&runtime.Lines, i18n.T("inspect.shell"), strings.Join(data.Runtime.Shell, " "))
	appendValue(&runtime.Lines, i18n.T("inspect.on_build"), strings.Join(data.Runtime.OnBuild, ", "))
	appendValue(&runtime.Lines, i18n.T("inspect.exposed_ports"), strings.Join(data.Runtime.ExposedPorts, ", "))
	appendValue(&runtime.Lines, i18n.T("inspect.volumes"), strings.Join(data.Runtime.Volumes, ", "))
	if len(data.Runtime.Environment) > 0 {
		runtime.Lines = append(runtime.Lines, i18n.T("inspect.environment")+":")
		runtime.Lines = append(runtime.Lines, data.Runtime.Environment...)
	}
	if len(data.Runtime.Healthcheck) > 0 {
		runtime.Lines = append(runtime.Lines, i18n.T("inspect.healthcheck")+":")
		runtime.Lines = append(runtime.Lines, data.Runtime.Healthcheck...)
	}
	if len(runtime.Lines) > 0 {
		sections = append(sections, runtime)
	}

	storage := detailSection{Title: i18n.T("inspect.section_storage")}
	appendValue(&storage.Lines, i18n.T("inspect.driver"), data.Driver)
	if data.LayerCount > 0 {
		appendValue(&storage.Lines, i18n.T("inspect.layers"), fmt.Sprintf("%d", data.LayerCount))
	}
	if len(storage.Lines) > 0 {
		sections = append(sections, storage)
	}

	if len(data.Labels) > 0 {
		labels := detailSection{Title: i18n.T("inspect.section_tags")}
		keys := make([]string, 0, len(data.Labels))
		for key := range data.Labels {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			labels.Lines = append(labels.Lines, key+"="+data.Labels[key])
		}
		sections = append(sections, labels)
	}

	history := detailSection{Title: i18n.T("inspect.section_history")}
	if data.IsManifest {
		history.Subtitle = i18n.T("inspect.history_variants")
		for _, variant := range data.ManifestVariants {
			platform := variant.Platform.OS + "/" + variant.Platform.Architecture
			if variant.Platform.Variant != "" {
				platform += "/" + variant.Platform.Variant
			}
			availability := "remote"
			if variant.Available {
				availability = "available"
			}
			history.Lines = append(history.Lines, fmt.Sprintf("%s: %s, %s, %s", platform, variant.Digest, utils.FormatBytes(float64(variant.Size)), availability))
		}
		if len(data.ManifestVariants) == 0 && data.HistoryError == "" {
			history.Lines = append(history.Lines, i18n.T("inspect.history_no_variants"))
		}
	} else {
		history.Subtitle = i18n.T("inspect.history_layers")
		for index, layer := range data.History {
			created := ""
			if layer.Created > 0 {
				created = time.Unix(layer.Created, 0).Format(time.RFC3339)
			}
			parts := []string{utils.FormatBytes(float64(layer.Size))}
			if created != "" {
				parts = append(parts, created)
			}
			if layer.CreatedBy != "" {
				parts = append(parts, layer.CreatedBy)
			}
			if layer.Comment != "" {
				parts = append(parts, i18n.T("inspect.comment")+": "+layer.Comment)
			}
			history.Lines = append(history.Lines, fmt.Sprintf("%d: %s", index+1, strings.Join(parts, " | ")))
		}
		if len(data.History) == 0 && data.HistoryError == "" {
			if data.HistorySource == runtimeapi.ImageHistoryPending {
				history.Lines = append(history.Lines, i18n.T("inspect.history_loading"))
			} else {
				history.Lines = append(history.Lines, i18n.T("inspect.history_empty"))
			}
		}
	}
	if data.HistoryError != "" {
		history.Lines = append(history.Lines, i18n.T("inspect.history_unavailable")+": "+data.HistoryError)
	}
	sections = append(sections, history)
	return sections
}

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

	sections[4].Lines = append(sections[4].Lines, labels...)

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
