package detail

// imageDetailSection identifies a parser state in buildImageDetailSections.
// The values mirror the section names shown in inspect.section_* i18n keys
// (summary, system, config, storage) plus a few intermediate bucket names
// (env, volumes, health, labels) used only while streaming content.
type imageDetailSection string

const (
	imgSectionSummary imageDetailSection = "summary"
	imgSectionSystem  imageDetailSection = "system"
	imgSectionConfig  imageDetailSection = "config"
	imgSectionStorage imageDetailSection = "storage"
	imgSectionVolumes imageDetailSection = "volumes"
	imgSectionHealth  imageDetailSection = "health"
	imgSectionLabels  imageDetailSection = "labels"
	imgSectionEnv     imageDetailSection = "env"
)

// Section-header delimiters as they appear in the parsed image detail
// text. Two delimiters may target the same parser section — e.g.
// "── Runtime ──" and "── Entrypoint / Cmd ──" both feed the config
// bucket — which is why classifyImageHeader returns a single canonical
// section rather than mapping 1:1.
const (
	imgHeaderSystem     = "── System ──"
	imgHeaderRuntime    = "── Runtime ──"
	imgHeaderEntrypoint = "── Entrypoint / Cmd ──"
	imgHeaderVolumes    = "── Volumes ──"
	imgHeaderHealth     = "── Healthcheck ──"
	imgHeaderStorage    = "── Storage ──"
	imgHeaderLabels     = "── Labels ──"
)

// Parametrized "── Environment (N vars) ──" delimiter; the parser
// matches it by prefix+suffix instead of an exact string.
const (
	imgEnvironmentHeaderPrefix = "── Environment ("
	imgEnvironmentHeaderSuffix = " vars) ──"
)

// classifyImageHeader maps a literal header line to its parser section,
// returning ok=false when the line isn't a known section marker. The
// runtime and entrypoint delimiters both feed the config bucket.
func classifyImageHeader(line string) (imageDetailSection, bool) {
	switch line {
	case imgHeaderSystem:
		return imgSectionSystem, true
	case imgHeaderRuntime, imgHeaderEntrypoint:
		return imgSectionConfig, true
	case imgHeaderVolumes:
		return imgSectionVolumes, true
	case imgHeaderHealth:
		return imgSectionHealth, true
	case imgHeaderStorage:
		return imgSectionStorage, true
	case imgHeaderLabels:
		return imgSectionLabels, true
	}
	return "", false
}
