package compose

import "github.com/elizabevil/docker-tui/internal/tui/tables"

func ProjectTableColumns() []tables.ColumnDef { return tc.Columns.Get("default") }

func ServiceTableColumns() []tables.ColumnDef { return tc.Columns.Get("services_sub") }
