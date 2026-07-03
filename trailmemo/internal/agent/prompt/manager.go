package prompt

import (
	"embed"
	"fmt"
	"strings"
	"sync"
)

// Manager loads and renders prompt templates with variable substitution.
// Templates are embedded from the templates/ directory at build time.
type Manager struct {
	mu     sync.RWMutex
	cache  map[string]*Template
	loader Loader
}

// Loader abstracts how to read template bytes. Use EmbedFS for production.
type Loader interface {
	Load(name string) ([]byte, error)
}

// NewManager creates a prompt manager with the given loader.
func NewManager(loader Loader) *Manager {
	return &Manager{
		cache:  make(map[string]*Template),
		loader: loader,
	}
}

// Render loads (or fetches from cache) the template and renders it with the given variables.
// 模板名称（name）和变量映射（vars）作为参数，返回渲染后的字符串（content）。
func (m *Manager) Render(name string, vars map[string]string) (string, error) {
	tmpl, err := m.get(name)
	if err != nil {
		return "", err
	}
	return tmpl.Render(vars), nil
}

// get loads a template by name, caching it after the first load.
// 作用：根据模板名称（name）加载并缓存模板（tmpl）。
// 参数：模板名称（name）作为参数，返回模板（tmpl）和错误（err）。
func (m *Manager) get(name string) (*Template, error) {
	m.mu.RLock()
	if tmpl, ok := m.cache[name]; ok {
		m.mu.RUnlock()
		return tmpl, nil
	}
	m.mu.RUnlock()

	raw, err := m.loader.Load(name) // 加载模板（name）的原始内容（raw）
	if err != nil {
		return nil, fmt.Errorf("prompt: load %s: %w", name, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock.
	if tmpl, ok := m.cache[name]; ok {
		return tmpl, nil
	}

	tmpl := NewTemplate(name, string(raw)) // 创建模板（tmpl）
	m.cache[name] = tmpl
	return tmpl, nil
}

// EmbedFS exposes Go embed.FS as a Loader.
type EmbedFS struct {
	FS embed.FS
}

func (e *EmbedFS) Load(name string) ([]byte, error) {
	return e.FS.ReadFile("templates/" + name + ".md")
}

// MapLoader is a simple in-memory loader, useful for tests.
type MapLoader map[string]string

func (m MapLoader) Load(name string) ([]byte, error) {
	raw, ok := m[name]
	if !ok {
		return nil, fmt.Errorf("template %s not found", name)
	}
	return []byte(raw), nil
}

// ── Variable helpers ─────────────────────────────

func (m *Manager) RenderSystem(name string, vars map[string]string) (*Message, error) {
	content, err := m.Render(name, vars)
	if err != nil {
		return nil, err
	}
	return &Message{Role: "system", Content: content}, nil
}

func (m *Manager) RenderUser(name string, vars map[string]string) (*Message, error) {
	content, err := m.Render(name, vars)
	if err != nil {
		return nil, err
	}
	return &Message{Role: "user", Content: content}, nil
}

// Message matches llm.Message so the prompt package doesn't need to import llm.
// At usage sites, convert to llm.Message with the same struct shape.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Template wraps a named template with {{variable}} substitution.
type Template struct {
	Name string // 模板名称
	Raw  string // 原始模板内容
}

func NewTemplate(name, raw string) *Template {
	return &Template{Name: name, Raw: raw}
}

// Render renders the template with the given variables.
// 作用：根据变量映射（vars）渲染模板（t）。
// 参数：变量映射（vars）作为参数，返回渲染后的字符串（result）。
func (t *Template) Render(vars map[string]string) string {
	result := t.Raw
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}

func (t *Template) Version() string {
	return "v1"
}
