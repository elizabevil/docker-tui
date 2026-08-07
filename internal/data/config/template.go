package config

import (
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Template renders the complete documented configuration template based on
// the embedded default values. Every section and field carries a Chinese
// comment; omitted keys fall back to the embedded defaults at load time.
// The output is a single YAML document that strictly decodes into UserConfig.
func Template() ([]byte, error) {
	app := AppPatch{}
	for _, name := range embeddedDefaultFiles {
		data, err := fs.ReadFile(DefaultsFS(), name)
		if err != nil {
			return nil, fmt.Errorf(errReadEmbeddedDefaultFormat, name, err)
		}
		var patch AppPatch
		if err := decodeJSONCStrict(data, &patch); err != nil {
			return nil, fmt.Errorf(errParseEmbeddedDefaultFormat, name, err)
		}
		if err := validateEmbeddedDefaultScope(name, patch); err != nil {
			return nil, err
		}
		mergeEmbeddedPatch(&app, patch)
	}

	var buf strings.Builder
	buf.WriteString(templateHeader)
	buf.WriteString(newline)
	writeScalar(&buf, emptyValue, keyVersion, descVersion, strconv.Itoa(CurrentConfigVersion))
	buf.WriteString(newline)
	writeContainer(&buf, emptyValue, keyApp, descApp)
	renderGeneral(&buf, indentUnit, app.General)
	buf.WriteString(newline)
	renderUI(&buf, indentUnit, app.UI)
	buf.WriteString(newline)
	renderDocker(&buf, indentUnit, app.Docker)
	buf.WriteString(newline)
	if err := renderRuntime(&buf, indentUnit, app.Runtime); err != nil {
		return nil, err
	}
	buf.WriteString(newline)
	renderKeymap(&buf, indentUnit, app.Keymap)
	buf.WriteString(newline)
	renderLogs(&buf, indentUnit, app.Logs)
	buf.WriteString(newline)
	renderLayout(&buf, indentUnit, app.Layout)
	buf.WriteString(newline)
	renderCommands(&buf, indentUnit, app.Commands)
	buf.WriteString(newline)
	writeSection(&buf, emptyValue, keyAppearance)
	writeScalar(&buf, indentUnit, keyTheme, descTheme, strconv.Quote(themeDefaultName))
	writeScalar(&buf, indentUnit, keyOverrides, descOverrides, emptyMapLiteral)
	return []byte(buf.String()), nil
}

// mergeEmbeddedPatch folds the single-section patch decoded from one embedded
// default file into the accumulated template patch.
func mergeEmbeddedPatch(target *AppPatch, source AppPatch) {
	if source.General != nil {
		target.General = source.General
	}
	if source.UI != nil {
		target.UI = source.UI
	}
	if source.Docker != nil {
		target.Docker = source.Docker
	}
	if source.Runtime != nil {
		target.Runtime = source.Runtime
	}
	if source.Keymap != nil {
		target.Keymap = source.Keymap
	}
	if source.Logs != nil {
		target.Logs = source.Logs
	}
	if source.Layout != nil {
		target.Layout = source.Layout
	}
	if source.Commands != nil {
		target.Commands = source.Commands
	}
}

func writeScalar(buf *strings.Builder, indent, key, desc, value string) {
	fmt.Fprintf(buf, scalarCommentFmt, indent, key, desc, value)
	fmt.Fprintf(buf, kvFmt, indent, key, value)
}

func writeContainer(buf *strings.Builder, indent, key, desc string) {
	fmt.Fprintf(buf, containerCommentFmt, indent, key, desc)
	fmt.Fprintf(buf, keyOnlyFmt, indent, key)
}

func writeSection(buf *strings.Builder, indent, key string) {
	fmt.Fprintf(buf, sectionHeaderFmt, indent, key)
	fmt.Fprintf(buf, keyOnlyFmt, indent, key)
}

func renderInt(buf *strings.Builder, indent, key, desc string, v *int) {
	if v == nil {
		return
	}
	writeScalar(buf, indent, key, desc, strconv.Itoa(*v))
}

func renderBool(buf *strings.Builder, indent, key, desc string, v *bool) {
	if v == nil {
		return
	}
	writeScalar(buf, indent, key, desc, strconv.FormatBool(*v))
}

func renderStringValue[T ~string](buf *strings.Builder, indent, key, desc string, v *T) {
	if v == nil {
		return
	}
	writeScalar(buf, indent, key, desc, strconv.Quote(string(*v)))
}

func renderDuration(buf *strings.Builder, indent, key, desc string, v *time.Duration) {
	if v == nil {
		return
	}
	writeScalar(buf, indent, key, desc, strconv.Quote(v.String()))
}

func renderBinding(buf *strings.Builder, indent, key, desc string, b *KeyBinding) {
	if b == nil {
		return
	}
	defaults := string(b.Primary)
	if b.Secondary != KeyEmpty {
		defaults += bindingSeparator + string(b.Secondary)
	}
	fmt.Fprintf(buf, scalarCommentFmt, indent, key, desc, strconv.Quote(defaults))
	fmt.Fprintf(buf, keyOnlyFmt, indent, key)
	fmt.Fprintf(buf, kvFmt, indent+indentUnit, keyPrimary, strconv.Quote(string(b.Primary)))
	if b.Secondary != KeyEmpty {
		fmt.Fprintf(buf, kvFmt, indent+indentUnit, keySecondary, strconv.Quote(string(b.Secondary)))
	}
}

func renderConnections(buf *strings.Builder, indent string, p *[]RuntimeConnection) error {
	if p == nil {
		return nil
	}
	if len(*p) == 0 {
		writeScalar(buf, indent, keyConnections, descConnections, emptyListLiteral)
		return nil
	}
	data, err := yaml.Marshal(*p)
	if err != nil {
		return err
	}
	writeContainer(buf, indent, keyConnections, descConnections)
	trimmed := strings.TrimSuffix(string(data), newline)
	childIndent := indent + indentUnit
	buf.WriteString(childIndent + strings.ReplaceAll(trimmed, newline, newline+childIndent))
	buf.WriteString(newline)
	return nil
}

func renderGeneral(buf *strings.Builder, indent string, p *GeneralPatch) {
	if p == nil {
		return
	}
	writeSection(buf, indent, keyGeneral)
	fi := indent + indentUnit
	renderInt(buf, fi, keyScrollHeight, descScrollHeight, p.ScrollHeight)
	renderStringValue(buf, fi, keyReporting, descReporting, p.Reporting)
	renderStringValue(buf, fi, keyLang, descLang, p.Lang)
	renderStringValue(buf, fi, keySizeFormat, descSizeFormat, p.SizeFormat)
}

func renderUI(buf *strings.Builder, indent string, p *UIPatch) {
	if p == nil {
		return
	}
	writeSection(buf, indent, keyUI)
	fi := indent + indentUnit
	renderBool(buf, fi, keyWrapMainPanel, descWrapMainPanel, p.WrapMainPanel)
	renderBool(buf, fi, keyReturnImmediately, descReturnImmediately, p.ReturnImmediately)
	renderBool(buf, fi, keyShowHelp, descShowHelp, p.ShowHelp)
	renderInt(buf, fi, keyHintTimeout, descHintTimeout, p.HintTimeout)
	renderBool(buf, fi, keyEnableMouse, descEnableMouse, p.EnableMouse)
	renderDialog(buf, fi, p.Dialog)
	renderWindow(buf, fi, p.Window)
	renderHeader(buf, fi, p.Header)
	renderTable(buf, fi, p.Table)
}

func renderDialog(buf *strings.Builder, indent string, p *DialogLayoutPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, keyDialog, descDialog)
	fi := indent + indentUnit
	renderResponsiveSize(buf, fi, keyWidth, descDialogWidth, p.Width)
	renderResponsiveSize(buf, fi, keyHeight, descDialogHeight, p.Height)
	renderDialogPosition(buf, fi, keyPosition, descDialogPosition, p.Position)
	renderPanelSize(buf, fi, keyPanelSize, descPanelSize, p.PanelSize)
}

func renderResponsiveSize(buf *strings.Builder, indent, key, desc string, p *ResponsiveSizePatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderInt(buf, fi, keyPercent, descPercent, p.Percent)
	renderInt(buf, fi, keyMin, descMin, p.Min)
	renderInt(buf, fi, keyMax, descMax, p.Max)
}

func renderDialogPosition(buf *strings.Builder, indent, key, desc string, p *DialogPositionPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderStringValue(buf, fi, keyHorizontal, descHorizontal, p.Horizontal)
	renderStringValue(buf, fi, keyVertical, descVertical, p.Vertical)
	renderInt(buf, fi, keyOffsetX, descOffsetX, p.OffsetX)
	renderInt(buf, fi, keyOffsetY, descOffsetY, p.OffsetY)
}

func renderPanelSize(buf *strings.Builder, indent, key, desc string, p *PanelSizePatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderInt(buf, fi, keyWidthPercent, descWidthPercent, p.WidthPercent)
	renderInt(buf, fi, keyHeightPercent, descHeightPercent, p.HeightPercent)
	renderInt(buf, fi, keyMaxWidth, descMaxWidth, p.MaxWidth)
	renderInt(buf, fi, keyMaxHeight, descMaxHeight, p.MaxHeight)
}

func renderWindow(buf *strings.Builder, indent string, p *WindowLayoutPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, keyWindow, descWindow)
	fi := indent + indentUnit
	renderInt(buf, fi, keyMarginTopPercent, descMarginTopPercent, p.MarginTopPercent)
	renderInt(buf, fi, keyMarginBottomPercent, descMarginBottomPercent, p.MarginBottomPercent)
	renderInt(buf, fi, keyContentWidthPercent, descContentWidthPercent, p.ContentWidthPercent)
}

func renderHeader(buf *strings.Builder, indent string, p *HeaderLayoutPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, keyHeader, descHeader)
	fi := indent + indentUnit
	renderHeaderColumns(buf, fi, keyColumns, descColumns, p.Columns)
	renderInt(buf, fi, keyKeystrokeContentRatio, descKeystrokeContentRatio, p.KeystrokeContentRatio)
}

func renderHeaderColumns(buf *strings.Builder, indent, key, desc string, p *HeaderColumnWeightsPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderInt(buf, fi, keyHost, descHost, p.Host)
	renderInt(buf, fi, keyConnection, descConnection, p.Connection)
	renderInt(buf, fi, keyKeystroke, descKeystroke, p.Keystroke)
	renderInt(buf, fi, keyLogo, descLogo, p.Logo)
}

func renderTable(buf *strings.Builder, indent string, p *TableLayoutPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, keyTable, descTable)
	fi := indent + indentUnit
	renderInt(buf, fi, keyColumnSpacing, descColumnSpacing, p.ColumnSpacing)
	renderStringValue(buf, fi, keyRowPrefix, descRowPrefix, p.RowPrefix)
	renderStringValue(buf, fi, keyRowPrefixSelected, descRowPrefixSelected, p.RowPrefixSelected)
	renderInt(buf, fi, keyRowSpacing, descRowSpacing, p.RowSpacing)
	renderSelectionInfo(buf, fi, keySelectionInfo, descSelectionInfo, p.SelectionInfo)
}

func renderSelectionInfo(buf *strings.Builder, indent, key, desc string, p *SelectionInfoPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderBool(buf, fi, keyEnabled, descEnabled, p.Enabled)
	renderInt(buf, fi, keyPadLines, descPadLines, p.PadLines)
}

func renderDocker(buf *strings.Builder, indent string, p *DockerPatch) {
	if p == nil {
		return
	}
	writeSection(buf, indent, keyDocker)
	fi := indent + indentUnit
	renderDuration(buf, fi, keyTimeout, descTimeout, p.Timeout)
	renderInt(buf, fi, keyStatsPollSec, descStatsPollSec, p.StatsPollSec)
}

func renderRuntime(buf *strings.Builder, indent string, p *RuntimePatch) error {
	if p == nil {
		return nil
	}
	writeSection(buf, indent, keyRuntime)
	fi := indent + indentUnit
	renderStringValue(buf, fi, keyDefault, descDefault, p.Default)
	renderRuntimeDiscovery(buf, fi, keyDiscovery, descDiscovery, p.Discovery)
	renderRuntimeHealth(buf, fi, keyHealth, descHealth, p.Health)
	return renderConnections(buf, fi, p.Connections)
}

func renderRuntimeDiscovery(buf *strings.Builder, indent, key, desc string, p *RuntimeDiscoveryPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderBool(buf, fi, keyLocalDocker, descLocalDocker, p.LocalDocker)
	renderBool(buf, fi, keyLocalPodman, descLocalPodman, p.LocalPodman)
}

func renderRuntimeHealth(buf *strings.Builder, indent, key, desc string, p *RuntimeHealthPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderInt(buf, fi, keyIntervalSec, descIntervalSec, p.IntervalSec)
	renderInt(buf, fi, keyTimeoutSec, descTimeoutSec, p.TimeoutSec)
	renderInt(buf, fi, keyFailureThreshold, descFailureThreshold, p.FailureThreshold)
}

func renderKeymap(buf *strings.Builder, indent string, p *KeymapPatch) {
	if p == nil {
		return
	}
	writeSection(buf, indent, keyKeymap)
	fi := indent + indentUnit
	renderGlobalKeymap(buf, fi, keyGlobal, descGlobal, p.Global)
	renderContainerKeymap(buf, fi, keyContainer, descContainer, p.Container)
	renderImageKeymap(buf, fi, keyImage, descImage, p.Image)
	renderResourceKeymap(buf, fi, keyVolume, descVolume, p.Volume)
	renderResourceKeymap(buf, fi, keyNetwork, descNetwork, p.Network)
	renderNavigationKeymap(buf, fi, keyNavigation, descNavigation, p.Navigation)
	renderDialogKeymap(buf, fi, keyDialog, descKeymapDialog, p.Dialog)
}

func renderGlobalKeymap(buf *strings.Builder, indent, key, desc string, p *GlobalKeymapPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderBinding(buf, fi, keyQuit, descQuit, p.Quit)
	renderBinding(buf, fi, keyActionBar, descActionBar, p.ActionBar)
	renderBinding(buf, fi, keyHelp, descHelp, p.Help)
	renderBinding(buf, fi, keyFilter, descFilter, p.Filter)
	renderBinding(buf, fi, keyRefresh, descRefresh, p.Refresh)
}

func renderContainerKeymap(buf *strings.Builder, indent, key, desc string, p *ContainerKeymapPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderBinding(buf, fi, keyStart, descStart, p.Start)
	renderBinding(buf, fi, keyStop, descStop, p.Stop)
	renderBinding(buf, fi, keyRestart, descRestart, p.Restart)
	renderBinding(buf, fi, keyKill, descKill, p.Kill)
	renderBinding(buf, fi, keyRemove, descRemove, p.Remove)
	renderBinding(buf, fi, keyLogs, descLogs, p.Logs)
	renderBinding(buf, fi, keyExec, descExec, p.Exec)
	renderBinding(buf, fi, keyInspect, descInspect, p.Inspect)
	renderBinding(buf, fi, keyStats, descStats, p.Stats)
	renderBinding(buf, fi, keyPause, descPause, p.Pause)
}

func renderImageKeymap(buf *strings.Builder, indent, key, desc string, p *ImageKeymapPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderBinding(buf, fi, keyPull, descPull, p.Pull)
	renderBinding(buf, fi, keyRemove, descRemove, p.Remove)
	renderBinding(buf, fi, keyPrune, descPrune, p.Prune)
	renderBinding(buf, fi, keyTag, descTag, p.Tag)
	renderBinding(buf, fi, keyPush, descPush, p.Push)
	renderBinding(buf, fi, keySave, descSave, p.Save)
	renderBinding(buf, fi, keyLoad, descLoad, p.Load)
	renderBinding(buf, fi, keyHistory, descHistory, p.History)
}

func renderResourceKeymap(buf *strings.Builder, indent, key, desc string, p *ResourceKeymapPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderBinding(buf, fi, keyCreate, descCreate, p.Create)
	renderBinding(buf, fi, keyPrune, descPrune, p.Prune)
	renderBinding(buf, fi, keyRemove, descRemove, p.Remove)
}

func renderNavigationKeymap(buf *strings.Builder, indent, key, desc string, p *NavigationKeymapPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderBinding(buf, fi, keyTabNext, descTabNext, p.TabNext)
	renderBinding(buf, fi, keyTabPrev, descTabPrev, p.TabPrev)
	renderBinding(buf, fi, keyUp, descUp, p.Up)
	renderBinding(buf, fi, keyDown, descDown, p.Down)
	renderBinding(buf, fi, keyEnter, descEnter, p.Enter)
	renderBinding(buf, fi, keyBack, descBack, p.Back)
	renderBinding(buf, fi, keyDelete, descDelete, p.Delete)
}

func renderDialogKeymap(buf *strings.Builder, indent, key, desc string, p *DialogKeymapPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderBinding(buf, fi, keyConfirm, descConfirm, p.Confirm)
	renderBinding(buf, fi, keyCancel, descCancel, p.Cancel)
}

func renderLogs(buf *strings.Builder, indent string, p *LogsPatch) {
	if p == nil {
		return
	}
	writeSection(buf, indent, keyLogs)
	fi := indent + indentUnit
	renderStringValue(buf, fi, keySince, descSince, p.Since)
	renderStringValue(buf, fi, keyTail, descTail, p.Tail)
	renderBool(buf, fi, keyTimestamps, descTimestamps, p.Timestamps)
}

func renderLayout(buf *strings.Builder, indent string, p *LayoutPatch) {
	if p == nil {
		return
	}
	writeSection(buf, indent, keyLayout)
	fi := indent + indentUnit
	renderSectionWeights(buf, fi, keySectionWeights, descSectionWeights, p.SectionWeights)
	renderBackground(buf, fi, keyBackground, descBackground, p.Background)
}

func renderSectionWeights(buf *strings.Builder, indent, key, desc string, p *SectionWeightsPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderInt(buf, fi, keyTop, descTop, p.Top)
	renderInt(buf, fi, keyContent, descContent, p.Content)
	renderInt(buf, fi, keyBottom, descBottom, p.Bottom)
}

func renderBackground(buf *strings.Builder, indent, key, desc string, p *BackgroundPatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderBool(buf, fi, keyEnable, descBgEnable, p.Enable)
	renderBool(buf, fi, keyFallthrough, descBgFallthrough, p.Fallthrough)
	renderStringValue(buf, fi, keyType, descBgType, p.Type)
	renderStringValue(buf, fi, keyColor, descBgColor, p.Color)
	renderBackgroundImage(buf, fi, keyImage, descBgImage, p.Image)
}

func renderBackgroundImage(buf *strings.Builder, indent, key, desc string, p *BackgroundImagePatch) {
	if p == nil {
		return
	}
	writeContainer(buf, indent, key, desc)
	fi := indent + indentUnit
	renderStringValue(buf, fi, keySizing, descBgSizing, p.Sizing)
	renderStringValue(buf, fi, keyPosition, descBgPosition, p.Position)
	renderInt(buf, fi, keyOpacity, descBgOpacity, p.Opacity)
}

func renderCommands(buf *strings.Builder, indent string, p *CommandPatch) {
	if p == nil {
		return
	}
	writeSection(buf, indent, keyCommands)
	fi := indent + indentUnit
	if p.DockerCompose != nil {
		writeContainer(buf, fi, keyDockerCompose, descDockerCompose)
		gi := fi + indentUnit
		renderStringValue(buf, gi, keyExecutable, descExecutable, p.DockerCompose.Executable)
		renderStringValue(buf, gi, keySubcommand, descSubcommand, p.DockerCompose.Subcommand)
	}
}
