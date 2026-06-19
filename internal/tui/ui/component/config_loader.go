package component

import (
	"encoding/json"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

// ConfigLoader loads embedded JSONC config into a strongly-typed struct.
// 使用 struct + method 绑定配置加载流程，避免各页面重复独立函数逻辑。
type ConfigLoader[T any] struct {
	RawData   []byte
	Fallback  T
	Normalize func(*T)
}

// Load parses JSONC data and returns validated config.
// On parse failure, fallback is returned.
func (l ConfigLoader[T]) Load() T {
	if len(l.RawData) == 0 {
		cfg := l.Fallback
		if l.Normalize != nil {
			l.Normalize(&cfg)
		}
		return cfg
	}

	clean := config.StripJSONComments(l.RawData)
	var cfg T
	if err := json.Unmarshal(clean, &cfg); err != nil {
		cfg = l.Fallback
	}
	if l.Normalize != nil {
		l.Normalize(&cfg)
	}
	return cfg
}
