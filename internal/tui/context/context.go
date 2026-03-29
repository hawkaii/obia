package context

import "github.com/hawkaii/obia/internal/config"

// ProgramContext holds shared state passed to all TUI components.
type ProgramContext struct {
	Config config.Config
	Width  int
	Height int
	Error  error
}

func New(cfg config.Config) *ProgramContext {
	return &ProgramContext{
		Config: cfg,
	}
}

func (ctx *ProgramContext) SetSize(w, h int) {
	ctx.Width = w
	ctx.Height = h
}

func (ctx *ProgramContext) VaultPath() string {
	return ctx.Config.Vault.Path
}

func (ctx *ProgramContext) DailyNotesFolder() string {
	return ctx.Config.Vault.DailyNotesFolder
}

func (ctx *ProgramContext) DailyNotesFormat() string {
	return ctx.Config.Vault.DailyNotesFormat
}
