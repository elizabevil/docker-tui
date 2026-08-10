package state

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// Path 字段需要区分两种文件系统语义（BR-041 §3）。
// dtui 本地路径指运行 dtui 进程的机器文件系统；
// 容器路径指选中容器内部路径。
type PathSource int

const (
	// PathLocal 是指运行 dtui 进程的机器文件系统。即便连接远程 daemon，
	// 本地目标仍写入 dtui 进程所在机器（BR-041 §3.1）。
	PathLocal PathSource = iota
	// PathContainer 是指选中容器内部路径（BR-041 §3.2）。
	PathContainer
)

// PathMode 决定候选来源与提交约束（BR-041 §5.2）。
type PathMode int

const (
	// PathAny 不限定文件或目录。
	PathAny PathMode = iota
	// PathFile 只允许文件。
	PathFile
	// PathDirectory 只允许目录。
	PathDirectory
	// PathSaveFile 是保存目标：最终文件名可以不存在，但父目录必须存在。
	PathSaveFile
)

// PathEntryType marks the kind of a path candidate for popup display (F/D/L
// prefix in the eza-l style listing).
type PathEntryType rune

const (
	PathEntryFile PathEntryType = 'F'
	PathEntryDir  PathEntryType = 'D'
	PathEntryLink PathEntryType = 'L'
)

// PathEntry is a single completion candidate. Detail fields (Mode/Owner/
// Group/Size/Mtime/LinkTarget) are populated by LocalPathProvider (via
// os.Lstat) and container completion (via ls -la); the popup shows them in
// eza-l column alignment.
type PathEntry struct {
	Name       string
	Path       string
	IsDir      bool
	Type       PathEntryType
	Mode       string // e.g. "drwxr-xr-x"
	Owner      string
	Group      string
	Size       int64
	Mtime      time.Time
	LinkTarget string // populated for symlinks (Type == L)
}

// IsHidden reports whether the entry is a dotfile/dotdir.
func (e PathEntry) IsHidden() bool {
	return strings.HasPrefix(e.Name, ".")
}

// PathCompletionRequest describes one completion request. ShowHidden opts in
// dotfiles regardless of the typed prefix.
type PathCompletionRequest struct {
	Path        string
	ContainerID string
	Mode        PathMode
	ShowHidden  bool
}

// ContainerPathCompleted returns an asynchronous container directory listing
// to the active form. Input and target identity prevent stale results from
// being applied after the user edits or closes the form.
type ContainerPathCompleted struct {
	ContainerID string
	FieldKey    string
	Input       string
	Entries     []PathEntry
	OpenPopup   bool
	Error       error
}

// PathProvider 是统一路径补全来源（BR-041 §6）。
// 第一阶段只实现 LocalPathProvider（容器相关留待第二阶段）。
type PathProvider interface {
	// Complete 返回与输入匹配的候选。不支持时返回 nil,err 以便上层保留手工输入。
	Complete(request PathCompletionRequest) ([]PathEntry, error)
}

// LocalPathProvider 补全 dtui 进程本地文件系统的路径（BR-041 §6.1）。
type LocalPathProvider struct {
	// CWD 是 dtui 启动时的当前工作目录，相对输入基于它展开。
	CWD string
}

// Complete 读取目录并返回候选项。目录优先、文件其次，同组按名称排序；
// 隐藏文件仅在 basename 以 '.' 开头或 request.ShowHidden 时显示；
// 每个条目通过 os.Lstat 填充 Mode/Owner/Group/Size/Mtime/Type；
// 文件系统错误返回 nil 以便手工输入。
func (p LocalPathProvider) Complete(request PathCompletionRequest) ([]PathEntry, error) {
	input := request.Path
	switch request.Mode {
	case PathFile, PathDirectory, PathAny, PathSaveFile:
	default:
		request.Mode = PathAny
	}

	home, _ := os.UserHomeDir()
	expanded, err := ExpandPath(input, home)
	if err != nil {
		return nil, err
	}
	// 显式位于目录内（以分隔符结尾或为空）时，列出该目录全部内容；
	// 否则把它当作待补全的名称，列出其父目录内容并按名称前缀过滤。
	trailingSep := strings.HasSuffix(expanded, "/") || strings.HasSuffix(expanded, `\`)
	abs := Absolute(expanded, p.CWD)
	dir := filepath.Dir(abs)
	prefixFile := ""
	if trailingSep {
		dir = abs
	} else {
		prefixFile = filepath.Base(abs)
	}
	if dir == "" {
		dir = "."
	}
	// File and save-file fields still need directory candidates so users can
	// traverse the filesystem before selecting the final file.
	showDirs := true
	showFiles := request.Mode != PathDirectory

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	needHidden := strings.HasPrefix(prefixFile, ".") || request.ShowHidden

	candidates := make([]PathEntry, 0, len(entries))
	for _, de := range entries {
		name := de.Name()
		if !needHidden && strings.HasPrefix(name, ".") {
			continue
		}
		isDir := de.IsDir()
		if isDir && !showDirs {
			continue
		}
		if !isDir && !showFiles {
			continue
		}
		if prefixFile != "" && !PrefixMatch(name, prefixFile) {
			continue
		}
		candidates = append(candidates, enrichLocalEntry(filepath.Join(dir, name), name, isDir))
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].IsDir != candidates[j].IsDir {
			return candidates[i].IsDir
		}
		return candidates[i].Name < candidates[j].Name
	})
	return candidates, nil
}

// PrefixMatch 判断 candidate 是否以 prefix 为前缀（忽略大小写，跨平台）。
func PrefixMatch(candidate, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(candidate), strings.ToLower(prefix))
}

// enrichLocalEntry populates Mode/Owner/Group/Size/Mtime/Type for a local
// path entry using os.Lstat. Symlinks are detected via mode bits and their
// target is resolved via os.Readlink.
func enrichLocalEntry(full, name string, isDir bool) PathEntry {
	entry := PathEntry{Name: name, Path: full, IsDir: isDir}
	info, err := os.Lstat(full)
	if err != nil {
		entry.Type = dirOrFileType(isDir)
		return entry
	}
	mode := info.Mode()
	entry.Mode = mode.String()
	entry.Size = info.Size()
	entry.Mtime = info.ModTime()
	isLink := mode&os.ModeSymlink != 0
	if isLink {
		if target, err := os.Readlink(full); err == nil {
			entry.LinkTarget = target
		}
		entry.Type = PathEntryLink
	} else {
		entry.Type = dirOrFileType(isDir)
	}
	return entry
}

func dirOrFileType(isDir bool) PathEntryType {
	if isDir {
		return PathEntryDir
	}
	return PathEntryFile
}

// JoinPath 规范化路径分隔符，兼容 Linux/macOS/Windows（BR-041 §6.1）。
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// ExpandPath 展开 ~ 与 $VAR 环境变量（BR-041 §3.1、§6.1）。纯函数。
func ExpandPath(input, envHome string) (string, error) {
	if input == "" {
		return input, nil
	}
	if input == "~" && envHome != "" {
		return expandEnv(envHome), nil
	}
	if envHome != "" && (strings.HasPrefix(input, "~/") || strings.HasPrefix(input, `~\`)) {
		return expandEnv(filepath.Join(envHome, input[2:])), nil
	}
	return expandEnv(input), nil
}

// Absolute 把 input 规整化为绝对路径（BR-041 §3.1）。
func Absolute(input, cwd string) string {
	if input == "" {
		return input
	}
	if filepath.IsAbs(input) {
		return filepath.Clean(input)
	}
	base := cwd
	if base == "" {
		base = "."
	}
	return filepath.Clean(JoinPath(base, input))
}

// expandEnv 展开 $VAR 与 ${VAR} 环境变量。
func expandEnv(s string) string {
	if !strings.ContainsRune(s, '$') {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		c := s[i]
		if c != '$' || i+1 >= len(s) {
			b.WriteByte(c)
			i++
			continue
		}
		switch {
		case s[i+1] == '{':
			end := strings.IndexByte(s[i+2:], '}')
			if end < 0 {
				b.WriteString(s[i:])
				return b.String()
			}
			name := s[i+2 : i+2+end]
			b.WriteString(os.Getenv(name))
			i += 2 + end + 1
		case isVarStart(s[i+1]):
			j := i + 1
			for j < len(s) && isVarPart(s[j]) {
				j++
			}
			name := s[i+1 : j]
			b.WriteString(os.Getenv(name))
			i = j
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

func isVarStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isVarPart(c byte) bool {
	return isVarStart(c) || (c >= '0' && c <= '9')
}

// SanitizeName 把 `容器名/源文件名清理成可嵌入文件名的片段：去前导 '/',
// 非法文件名字符替换为 '-'；空/无效时回退 fallback（空容器名回退 short ID）。
func SanitizeName(s, fallback string) string {
	if s == "" {
		return fallback
	}
	bad := func(r rune) bool {
		return r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' ||
			r == '<' || r == '>' || r == '|' || r == 0 || r < 32
	}
	s = strings.TrimLeft(s, "/")
	var b strings.Builder
	for _, r := range s {
		if bad(r) {
			b.WriteRune('-')
		} else {
			b.WriteRune(r)
		}
	}
	if out := b.String(); out != "" {
		return out
	}
	return fallback
}

// PathBase 返回路径的 basename；根目录返回 "rootfs"（BR-041 §4.1）。
func PathBase(p string) string {
	base := filepath.Base(strings.ReplaceAll(p, "\\", "/"))
	if base == "." || base == "/" || base == string(filepath.Separator) || base == "" {
		return "rootfs"
	}
	return base
}

// DefaultCopyName 生成 Copy 的默认本地目标名（BR-041 §4.1）:
//
//	<cwd>/<container-name>-<source-base>-<YYYYMMDD-HHMMSS>.tar
func DefaultCopyName(cwd, containerName, sourceBase, shortID string, t time.Time) string {
	ctr := SanitizeName(containerName, shortID)
	if ctr == "" {
		ctr = "rootfs"
	}
	return filepath.Join(cwd, ctr+"-"+PathBase(sourceBase)+"-"+t.Format(utils.FileNameDateTime)+".tar")
}

// DefaultExportName 生成 Export 的本地目标名（BR-041 §4.2）:
//
//	<cwd>/<container-name>-filesystem-<YYYYMMDD-HHMMSS>.tar
func DefaultExportName(cwd, containerName, shortID string, t time.Time) string {
	ctr := SanitizeName(containerName, shortID)
	if ctr == "" {
		ctr = "rootfs"
	}
	return filepath.Join(cwd, ctr+"-filesystem-"+t.Format(utils.FileNameDateTime)+".tar")
}

// DefaultImageSaveName generates an absolute local archive path for Image Save.
func DefaultImageSaveName(cwd, imageRef, shortID string, t time.Time) string {
	// Registry hosts and repository namespaces do not help identify the local
	// archive. Keep the final image name and tag, matching <image-name>-<tag>.
	if slash := strings.LastIndexAny(imageRef, `/\`); slash >= 0 {
		imageRef = imageRef[slash+1:]
	}
	name := SanitizeName(imageRef, shortID)
	if name == "" {
		name = "image"
	}
	return filepath.Join(cwd, name+"-"+t.Format(utils.FileNameDateTime)+".tar")
}
