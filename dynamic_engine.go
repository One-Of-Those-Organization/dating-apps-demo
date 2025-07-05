package main

import (
    "fmt"
    "html/template"
    "io"
    "os"
    "path/filepath"
    "sync"
    "time"
)

type DynamicEngine struct {
    directories []string
    extension   string
    funcMap     template.FuncMap
    left        string
    right       string
    mutex       sync.RWMutex
    cache       map[string]*template.Template
    cacheTime   map[string]time.Time
    cacheTTL    time.Duration
}

func NewDynamicEngine(directories []string, extension string) *DynamicEngine {
    return &DynamicEngine{
        directories: directories,
        extension: extension,
        funcMap:   make(template.FuncMap),
        left:      "{{",
        right:     "}}",
        cache:     make(map[string]*template.Template),
        cacheTime: make(map[string]time.Time),
        cacheTTL:  5 * time.Minute, // Cache templates for 5 minutes
    }
}

func (e *DynamicEngine) AddFunc(name string, fn interface{}) {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    e.funcMap[name] = fn
}

func (e *DynamicEngine) Load() error {
    // Required by Views interface but we'll load templates dynamically
    return nil
}

func (e *DynamicEngine) Render(out io.Writer, name string, binding interface{}, layout ...string) error {
    tmpl, err := e.getTemplate(name)
    if err != nil {
        return err
    }

    if len(layout) > 0 && layout[0] != "" {
        return e.renderWithLayout(out, name, binding, layout[0])
    }

    return tmpl.Execute(out, binding)
}

func (e *DynamicEngine) getTemplate(name string) (*template.Template, error) {
    e.mutex.RLock()

    // Check cache and TTL
    if tmpl, exists := e.cache[name]; exists {
        if time.Since(e.cacheTime[name]) < e.cacheTTL {
            e.mutex.RUnlock()
            return tmpl, nil
        }
    }
    e.mutex.RUnlock()

    // Load template from disk
    return e.loadTemplate(name)
}

func (e *DynamicEngine) loadTemplate(name string) (*template.Template, error) {
    e.mutex.Lock()
    defer e.mutex.Unlock()

    // Double-check after acquiring write lock
    if tmpl, exists := e.cache[name]; exists {
        if time.Since(e.cacheTime[name]) < e.cacheTTL {
            return tmpl, nil
        }
    }

    var templatePath string
    var found bool
    for _, dir := range e.directories {
        tryPath := filepath.Join(dir, name+e.extension)
        if _, err := os.Stat(tryPath); err == nil {
            templatePath = tryPath
            found = true
            break
        }
    }
    if !found {
        return nil, fmt.Errorf("template %s does not exist in any directories", name)
    }

    // Check if file exists
    if _, err := os.Stat(templatePath); os.IsNotExist(err) {
        return nil, fmt.Errorf("template %s does not exist", name)
    }

    // Read template file
    content, err := os.ReadFile(templatePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read template %s: %w", name, err)
    }

    // Create and parse template
    tmpl := template.New(name).Delims(e.left, e.right).Funcs(e.funcMap)
    tmpl, err = tmpl.Parse(string(content))
    if err != nil {
        return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
    }

    // Cache the template
    e.cache[name] = tmpl
    e.cacheTime[name] = time.Now()

    return tmpl, nil
}

func (e *DynamicEngine) renderWithLayout(out io.Writer, templateName string, binding interface{}, layoutName string) error {
    // Load both template and layout
    tmpl, err := e.getTemplate(templateName)
    if err != nil {
        return err
    }

    layout, err := e.getTemplate(layoutName)
    if err != nil {
        return err
    }

    // Create a new template that includes both
    combined := template.New("combined").Delims(e.left, e.right).Funcs(e.funcMap)

    // Add the main template
    combined, err = combined.AddParseTree(templateName, tmpl.Tree)
    if err != nil {
        return err
    }

    // Add the layout template
    combined, err = combined.AddParseTree(layoutName, layout.Tree)
    if err != nil {
        return err
    }

    return combined.ExecuteTemplate(out, layoutName, binding)
}

// ClearCache clears the template cache (useful for admin operations)
func (e *DynamicEngine) ClearCache() {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    e.cache = make(map[string]*template.Template)
    e.cacheTime = make(map[string]time.Time)
}

// ReloadTemplate forces reload of a specific template
func (e *DynamicEngine) ReloadTemplate(name string) error {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    delete(e.cache, name)
    delete(e.cacheTime, name)
    return nil
}
