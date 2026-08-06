package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/elizabevil/docker-tui/internal/utils"
)

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return emptyValue, fmt.Errorf(errGetHomeDirFormat, err)
	}
	return filepath.Join(home, userConfigRootName, configDirName), nil
}

func ConfigFile() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return emptyValue, err
	}
	return filepath.Join(dir, configFileName), nil
}

func ValidateApp(cfg *AppConfig) error {
	if cfg == nil {
		return errors.New(errAppConfigNil)
	}
	if cfg.Version != CurrentConfigVersion {
		return fmt.Errorf(errVersionFormat, CurrentConfigVersion)
	}
	if cfg.General.ScrollHeight < 0 {
		return errors.New(errScrollHeightNegative)
	}
	if cfg.General.SizeFormat != SizeFormatBinary && cfg.General.SizeFormat != SizeFormatSI {
		return errors.New(errSizeFormatInvalid)
	}
	if cfg.General.Lang != LanguageEnglish && cfg.General.Lang != LanguageChinese {
		return errors.New(errLanguageInvalid)
	}
	if cfg.General.Reporting != ReportingOff {
		return errors.New(errReportingModeInvalid)
	}
	if cfg.UI.HintTimeout < 0 {
		return errors.New(errHintTimeoutNegative)
	}
	if err := cfg.UI.Dialog.Width.Validate(); err != nil {
		return fmt.Errorf(errValidateAtFormat, pathDialogWidth, err)
	}
	if err := cfg.UI.Dialog.Height.Validate(); err != nil {
		return fmt.Errorf(errValidateAtFormat, pathDialogHeight, err)
	}
	if err := cfg.UI.Dialog.Position.Validate(); err != nil {
		return err
	}
	if err := cfg.UI.Dialog.PanelSize.Validate(); err != nil {
		return err
	}
	if cfg.UI.Window.MarginTopPercent < 0 || cfg.UI.Window.MarginTopPercent > 15 ||
		cfg.UI.Window.MarginBottomPercent < 0 || cfg.UI.Window.MarginBottomPercent > 15 {
		return errors.New(errWindowMarginInvalid)
	}
	if cfg.UI.Window.ContentWidthPercent <= 0 || cfg.UI.Window.ContentWidthPercent > 100 {
		return errors.New(errWindowWidthInvalid)
	}
	headerColumns := cfg.UI.Header.Columns
	if headerColumns.Host <= 0 || headerColumns.Connection <= 0 || headerColumns.Keystroke <= 0 || headerColumns.Logo <= 0 {
		return errors.New(errHeaderWeightInvalid)
	}
	if cfg.UI.Header.KeystrokeContentRatio <= 0 || cfg.UI.Header.KeystrokeContentRatio > 100 {
		return errors.New(errHeaderKeystrokeRatioInvalid)
	}
	if cfg.UI.Table.ColumnSpacing <= 0 {
		return errors.New(errTableColumnSpacingInvalid)
	}
	if cfg.UI.Table.RowPrefix == emptyValue || cfg.UI.Table.RowPrefixSelected == emptyValue {
		return errors.New(errTablePrefixRequired)
	}
	if cfg.UI.Table.RowSpacing < 0 {
		return errors.New(errTableRowSpacingInvalid)
	}
	if cfg.UI.Table.SelectionInfo.PadLines < 0 {
		return errors.New(errTableSelectionPaddingInvalid)
	}
	if err := cfg.Keymap.Validate(); err != nil {
		return err
	}
	if err := cfg.Runtime.Validate(); err != nil {
		return err
	}
	if err := cfg.Layout.Background.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Commands.DockerCompose.Executable) == emptyValue {
		return errors.New(errCommandExecutableRequired)
	}
	if strings.TrimSpace(cfg.Commands.DockerCompose.Subcommand) == emptyValue {
		return errors.New(errCommandSubcommandRequired)
	}
	return nil
}

// MaxWidth / MaxHeight are advisory caps; a zero value means "no cap"
// so the panel-sizing helpers fall back to cfg.Width / cfg.Height.
func (p PanelSize) Validate() error {
	if p.WidthPercent <= 0 || p.WidthPercent > 100 {
		return fmt.Errorf(errResponsivePercentFormat, pathPanelSizeWidth)
	}
	if p.HeightPercent <= 0 || p.HeightPercent > 100 {
		return fmt.Errorf(errResponsivePercentFormat, pathPanelSizeHeight)
	}
	return nil
}

// Validate reports whether the responsive size bounds are usable. The
// caller supplies any field-path context since ResponsiveSize does not
// own it.
func (r ResponsiveSize) Validate() error {
	if r.Percent <= 0 || r.Percent > 100 {
		return errors.New(errResponsivePercentInvalid)
	}
	if r.Min <= 0 || r.Max < r.Min {
		return errors.New(errResponsiveBoundsInvalid)
	}
	return nil
}

func (p DialogPosition) Validate() error {
	switch p.Horizontal {
	case HorizontalLeft, HorizontalCenter, HorizontalRight:
	default:
		return errors.New(errDialogHorizontalInvalid)
	}
	switch p.Vertical {
	case VerticalTop, VerticalCenter, VerticalBottom:
	default:
		return errors.New(errDialogVerticalInvalid)
	}
	return nil
}

func (b BackgroundConfig) Validate() error {
	switch b.Type {
	case BackgroundSolid, BackgroundGradient, BackgroundImageType:
	default:
		return errors.New(errBackgroundTypeInvalid)
	}
	if b.Image.Sizing != BackgroundSizing(emptyValue) &&
		b.Image.Sizing != BackgroundSizingCover && b.Image.Sizing != BackgroundSizingContain {
		return errors.New(errBackgroundSizingInvalid)
	}
	if b.Image.Position != BackgroundPosition(emptyValue) &&
		b.Image.Position != BackgroundPositionTop && b.Image.Position != BackgroundPositionCenter &&
		b.Image.Position != BackgroundPositionBottom {
		return errors.New(errBackgroundPositionInvalid)
	}
	return nil
}

type namedBinding struct {
	path    string
	binding KeyBinding
}

func (k KeymapConfig) Validate() error {
	bindings := []namedBinding{
		{pathKeymapGlobalQuit, k.Global.Quit}, {pathKeymapGlobalActionBar, k.Global.ActionBar},
		{pathKeymapGlobalHelp, k.Global.Help}, {pathKeymapGlobalFilter, k.Global.Filter}, {pathKeymapGlobalRefresh, k.Global.Refresh},
		{pathKeymapContainerStart, k.Container.Start}, {pathKeymapContainerStop, k.Container.Stop},
		{pathKeymapContainerRestart, k.Container.Restart}, {pathKeymapContainerKill, k.Container.Kill},
		{pathKeymapContainerRemove, k.Container.Remove}, {pathKeymapContainerLogs, k.Container.Logs},
		{pathKeymapContainerExec, k.Container.Exec}, {pathKeymapContainerInspect, k.Container.Inspect},
		{pathKeymapContainerStats, k.Container.Stats}, {pathKeymapContainerPause, k.Container.Pause},
		{pathKeymapContainerUpdate, k.Container.Update}, {pathKeymapContainerDiff, k.Container.Diff},
		{pathKeymapContainerExport, k.Container.Export}, {pathKeymapContainerCommit, k.Container.Commit},
		{pathKeymapContainerWait, k.Container.Wait}, {pathKeymapContainerCopy, k.Container.Copy},
		{pathKeymapImagePull, k.Image.Pull}, {pathKeymapImageRemove, k.Image.Remove},
		{pathKeymapImagePrune, k.Image.Prune}, {pathKeymapImageTag, k.Image.Tag},
		{pathKeymapImagePush, k.Image.Push}, {pathKeymapImageSave, k.Image.Save},
		{pathKeymapImageLoad, k.Image.Load}, {pathKeymapImageHistory, k.Image.History},
		{pathKeymapVolumeCreate, k.Volume.Create}, {pathKeymapVolumePrune, k.Volume.Prune}, {pathKeymapVolumeRemove, k.Volume.Remove},
		{pathKeymapNetworkCreate, k.Network.Create}, {pathKeymapNetworkPrune, k.Network.Prune}, {pathKeymapNetworkRemove, k.Network.Remove},
		{pathKeymapNavigationTabNext, k.Navigation.TabNext}, {pathKeymapNavigationTabPrev, k.Navigation.TabPrev},
		{pathKeymapNavigationUp, k.Navigation.Up}, {pathKeymapNavigationDown, k.Navigation.Down},
		{pathKeymapNavigationEnter, k.Navigation.Enter}, {pathKeymapNavigationBack, k.Navigation.Back},
		{pathKeymapNavigationDelete, k.Navigation.Delete}, {pathKeymapDialogConfirm, k.Dialog.Confirm},
		{pathKeymapDialogCancel, k.Dialog.Cancel},
	}
	for _, item := range bindings {
		primary := strings.TrimSpace(string(item.binding.Primary))
		secondary := strings.TrimSpace(string(item.binding.Secondary))
		if primary != string(item.binding.Primary) || secondary != string(item.binding.Secondary) {
			return fmt.Errorf(errKeysWhitespaceFormat, item.path)
		}
		if primary == emptyValue && secondary != emptyValue {
			return fmt.Errorf(errSecondaryRequiresPrimaryFormat, item.path)
		}
		if primary != emptyValue && strings.EqualFold(primary, secondary) {
			return fmt.Errorf(errDuplicateBindingFormat, item.path)
		}
	}
	return nil
}

func (r RuntimeConfig) Validate() error {
	if err := r.Health.Validate(); err != nil {
		return err
	}
	names := make(map[string]struct{}, len(r.Connections))
	for i, connection := range r.Connections {
		if err := connection.Validate(); err != nil {
			return fmt.Errorf(errRuntimeConnectionFormat, i, err)
		}
		if _, exists := names[connection.Name]; exists {
			return fmt.Errorf(errRuntimeConnectionDuplicateFormat, i, connection.Name)
		}
		names[connection.Name] = struct{}{}
	}
	if r.Default != DefaultConnectionLocalDocker && r.Default != DefaultConnectionLocalPodman {
		if _, exists := names[r.Default]; !exists {
			return fmt.Errorf(errRuntimeDefaultMissingFormat, r.Default)
		}
	}
	return nil
}

func (h RuntimeHealthConfig) Validate() error {
	if h.IntervalSec <= 0 || h.TimeoutSec <= 0 || h.FailureThreshold <= 0 {
		return errors.New(errRuntimeHealthInvalid)
	}
	return nil
}

func (c RuntimeConnection) Validate() error {
	if strings.TrimSpace(c.Name) == emptyValue {
		return errors.New(errConnectionNameRequired)
	}
	if c.Driver != RuntimeDriverDocker && c.Driver != RuntimeDriverPodman {
		return errors.New(errConnectionDriverInvalid)
	}
	if strings.TrimSpace(c.Endpoint) == emptyValue {
		return errors.New(errConnectionEndpointRequired)
	}
	endpoint, err := url.Parse(c.Endpoint)
	if err != nil || endpoint.Scheme == emptyValue {
		return errors.New(errConnectionEndpointInvalid)
	}
	if c.TLS.Enabled && endpoint.Scheme == EndpointSchemeUnix {
		return errors.New(errTLSUnixSocket)
	}
	return c.TLS.Validate()
}

func (tls RuntimeTLSConfig) Validate() error {
	if !tls.Enabled {
		return nil
	}
	if tls.Verify == tls.InsecureSkipVerify {
		return errors.New(errTLSVerifyConflict)
	}
	if tls.Verify && strings.TrimSpace(tls.CAFile) == emptyValue {
		return errors.New(errTLSCARequired)
	}
	if (strings.TrimSpace(tls.CertFile) == emptyValue) != (strings.TrimSpace(tls.KeyFile) == emptyValue) {
		return errors.New(errTLSClientPairRequired)
	}
	return nil
}

func ValidateTheme(theme *Theme) error {
	if theme == nil {
		return errors.New(errThemeNil)
	}
	colors := []Color{theme.Palette.Primary, theme.Palette.Success, theme.Palette.Warning, theme.Palette.Danger, theme.Palette.Info, theme.Palette.Accent, theme.Palette.AccentSecondary, theme.Palette.Foreground, theme.Palette.ForegroundMuted, theme.Palette.BackgroundSubtle, theme.Palette.BackgroundDeep, theme.Palette.Background}
	for _, color := range colors {
		if !isValidColor(string(color)) {
			return fmt.Errorf(errThemePaletteColorFormat, color)
		}
	}
	if isTransparentBase(theme.Palette.Background) {
		return fmt.Errorf(errThemePaletteTransparentBase, theme.Palette.Background)
	}
	if isTransparentBase(theme.Palette.Foreground) {
		return fmt.Errorf(errThemePaletteTransparentBase, theme.Palette.Foreground)
	}
	switch theme.Border.Kind {
	case BorderRounded, BorderSingle, BorderDouble, BorderThick, BorderHidden:
	default:
		return errors.New(errThemeBorderKindInvalid)
	}
	if theme.Dialog.OverlayOpacity > 100 {
		return errors.New(errThemeOverlayOpacityInvalid)
	}
	refs := []ColorRef{
		theme.Border.Active, theme.Border.Inactive, theme.Border.Focused,
		theme.Header.Background, theme.Header.Label, theme.Header.Value, theme.Header.Logo,
		theme.Main.BorderActive, theme.Main.BorderInactive, theme.Main.Title, theme.Main.TableHeader,
		theme.Main.RowSelected, theme.Main.RowText, theme.Main.Footer, theme.Main.PanelBackground,
		theme.Main.ActionBarBackground, theme.Main.MessageRailBackground, theme.Main.QueryBarBackground,
		theme.Footer.StatusBackground, theme.Footer.ShortcutBackground, theme.Footer.Key,
		theme.Footer.Description, theme.Footer.Separator,
		theme.Dialog.Border, theme.Dialog.Title, theme.Dialog.Body, theme.Dialog.BodyBackground,
		theme.Dialog.OptionActive, theme.Dialog.OptionInactive, theme.Dialog.Overlay,
		theme.Toast.Success, theme.Toast.Error, theme.Toast.Background,
		theme.Text.Info, theme.Text.Error, theme.Text.Success, theme.Text.Warning, theme.Text.Dim,
		theme.Text.HelpKey, theme.Text.HelpDescription,
	}
	for _, ref := range refs {
		if !isColorRef(ref) {
			return fmt.Errorf(errThemeColorReferenceFormat, ref)
		}
	}
	return nil
}

func isColorRef(ref ColorRef) bool {
	if (ref.Token == ColorToken(emptyValue)) == (ref.Value == Color(emptyValue)) {
		return false
	}
	if ref.Value != Color(emptyValue) {
		return isValidColor(string(ref.Value))
	}
	switch ref.Token {
	case ColorTokenPrimary, ColorTokenSuccess, ColorTokenWarning, ColorTokenDanger, ColorTokenInfo, ColorTokenAccent, ColorTokenAccentSecondary, ColorTokenForeground, ColorTokenForegroundMuted, ColorTokenBackground, ColorTokenBackgroundSubtle, ColorTokenBackgroundDeep, ColorTokenTransparent:
		return true
	default:
		return false
	}
}

// isValidColor reports whether value is a supported color format. It
// delegates to utils.ParseColor, which accepts palette names, standard
// CSS color names, #hex (3/4/6/8 digits), rgb()/rgba() functions, and
// ANSI 0-255 color indices — matching the formats the style layer can
// render (see design/current-design.md §3.2).
func isValidColor(value string) bool {
	_, ok := utils.ParseColor(value)
	return ok
}

// isTransparentBase reports whether value resolves to the V3 "transparent"
// sentinel via ParseColor. Palette.Background and Palette.Foreground act as
// blend bases for 0<α<255 colors and cannot themselves be transparent.
func isTransparentBase(value Color) bool {
	_, ok := utils.ParseColor(string(value))
	if !ok {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(string(value)), keywordTransparent)
}

const keywordTransparent = "transparent"
