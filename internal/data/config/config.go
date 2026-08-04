package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
	if err := validateResponsiveSize(pathDialogWidth, cfg.UI.Dialog.Width); err != nil {
		return err
	}
	if err := validateResponsiveSize(pathDialogHeight, cfg.UI.Dialog.Height); err != nil {
		return err
	}
	if err := validateDialogPosition(cfg.UI.Dialog.Position); err != nil {
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
	if err := validateKeymap(cfg.Keymap); err != nil {
		return err
	}
	if err := cfg.Runtime.Validate(); err != nil {
		return err
	}
	if err := validateBackground(cfg.Layout.Background); err != nil {
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

func validateBackground(background BackgroundConfig) error {
	switch background.Type {
	case BackgroundSolid, BackgroundGradient, BackgroundImageType:
	default:
		return errors.New(errBackgroundTypeInvalid)
	}
	if background.Image.Sizing != BackgroundSizing(emptyValue) &&
		background.Image.Sizing != BackgroundSizingCover && background.Image.Sizing != BackgroundSizingContain {
		return errors.New(errBackgroundSizingInvalid)
	}
	if background.Image.Position != BackgroundPosition(emptyValue) &&
		background.Image.Position != BackgroundPositionTop && background.Image.Position != BackgroundPositionCenter &&
		background.Image.Position != BackgroundPositionBottom {
		return errors.New(errBackgroundPositionInvalid)
	}
	return nil
}

func validateDialogPosition(position DialogPosition) error {
	switch position.Horizontal {
	case HorizontalLeft, HorizontalCenter, HorizontalRight:
	default:
		return errors.New(errDialogHorizontalInvalid)
	}
	switch position.Vertical {
	case VerticalTop, VerticalCenter, VerticalBottom:
	default:
		return errors.New(errDialogVerticalInvalid)
	}
	return nil
}

type namedBinding struct {
	path    string
	binding KeyBinding
}

func validateKeymap(keymap KeymapConfig) error {
	bindings := []namedBinding{
		{pathKeymapGlobalQuit, keymap.Global.Quit}, {pathKeymapGlobalActionBar, keymap.Global.ActionBar},
		{pathKeymapGlobalHelp, keymap.Global.Help}, {pathKeymapGlobalFilter, keymap.Global.Filter}, {pathKeymapGlobalRefresh, keymap.Global.Refresh},
		{pathKeymapContainerStart, keymap.Container.Start}, {pathKeymapContainerStop, keymap.Container.Stop},
		{pathKeymapContainerRestart, keymap.Container.Restart}, {pathKeymapContainerKill, keymap.Container.Kill},
		{pathKeymapContainerRemove, keymap.Container.Remove}, {pathKeymapContainerLogs, keymap.Container.Logs},
		{pathKeymapContainerExec, keymap.Container.Exec}, {pathKeymapContainerInspect, keymap.Container.Inspect},
		{pathKeymapContainerStats, keymap.Container.Stats}, {pathKeymapContainerPause, keymap.Container.Pause},
		{pathKeymapContainerUpdate, keymap.Container.Update}, {pathKeymapContainerDiff, keymap.Container.Diff},
		{pathKeymapContainerExport, keymap.Container.Export}, {pathKeymapContainerCommit, keymap.Container.Commit},
		{pathKeymapContainerWait, keymap.Container.Wait}, {pathKeymapContainerCopy, keymap.Container.Copy},
		{pathKeymapImagePull, keymap.Image.Pull}, {pathKeymapImageRemove, keymap.Image.Remove},
		{pathKeymapImagePrune, keymap.Image.Prune}, {pathKeymapImageTag, keymap.Image.Tag},
		{pathKeymapImagePush, keymap.Image.Push}, {pathKeymapImageSave, keymap.Image.Save},
		{pathKeymapImageLoad, keymap.Image.Load}, {pathKeymapImageHistory, keymap.Image.History},
		{pathKeymapVolumeCreate, keymap.Volume.Create}, {pathKeymapVolumePrune, keymap.Volume.Prune}, {pathKeymapVolumeRemove, keymap.Volume.Remove},
		{pathKeymapNetworkCreate, keymap.Network.Create}, {pathKeymapNetworkPrune, keymap.Network.Prune}, {pathKeymapNetworkRemove, keymap.Network.Remove},
		{pathKeymapNavigationTabNext, keymap.Navigation.TabNext}, {pathKeymapNavigationTabPrev, keymap.Navigation.TabPrev},
		{pathKeymapNavigationUp, keymap.Navigation.Up}, {pathKeymapNavigationDown, keymap.Navigation.Down},
		{pathKeymapNavigationEnter, keymap.Navigation.Enter}, {pathKeymapNavigationBack, keymap.Navigation.Back},
		{pathKeymapNavigationDelete, keymap.Navigation.Delete}, {pathKeymapDialogConfirm, keymap.Dialog.Confirm},
		{pathKeymapDialogCancel, keymap.Dialog.Cancel},
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

func validateResponsiveSize(path string, size ResponsiveSize) error {
	if size.Percent <= 0 || size.Percent > 100 {
		return fmt.Errorf(errResponsivePercentFormat, path)
	}
	if size.Min <= 0 || size.Max < size.Min {
		return fmt.Errorf(errResponsiveBoundsFormat, path)
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
		if !isHexColor(string(color)) {
			return fmt.Errorf(errThemePaletteColorFormat, color)
		}
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
		return isHexColor(string(ref.Value))
	}
	switch ref.Token {
	case ColorTokenPrimary, ColorTokenSuccess, ColorTokenWarning, ColorTokenDanger, ColorTokenInfo, ColorTokenAccent, ColorTokenAccentSecondary, ColorTokenForeground, ColorTokenForegroundMuted, ColorTokenBackground, ColorTokenBackgroundSubtle, ColorTokenBackgroundDeep, ColorTokenTransparent:
		return true
	default:
		return false
	}
}

func isHexColor(value string) bool {
	if len(value) != 4 && len(value) != 7 && len(value) != 9 {
		return false
	}
	if value[0] != '#' {
		return false
	}
	for _, char := range value[1:] {
		if char < '0' || char > '9' {
			lower := char | 0x20
			if lower < 'a' || lower > 'f' {
				return false
			}
		}
	}
	return true
}
