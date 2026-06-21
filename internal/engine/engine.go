package engine

// Engine 输入法引擎主循环（MVP 阶段由 Server 直接驱动）
type Engine struct {
	Speller  Speller
	Transltr *Translator
}

// NewEngine 创建输入法引擎
func NewEngine(speller Speller, translator *Translator) *Engine {
	return &Engine{
		Speller:  speller,
		Transltr: translator,
	}
}
