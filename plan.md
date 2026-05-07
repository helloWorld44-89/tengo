# Shippable CLI Text Editor Plan

## Goal
Deliver a cross-platform CLI text editor in Go with:
- Quick edit mode (current feature)
- Full editor mode with advanced shortcuts
- Syntax verification for YAML, JSON, XML, and INI

## 4-Phase Approach

### Phase A: Quick Editor Stabilization & UI Polish
- Refactor error handling for user-friendly messages (no panics) **[Done]**
- Add minimal tests for quick editor core functions **[Done]**
- Audit and finalize all keyboard shortcuts for consistency and discoverability
- Add a help popup or overlay listing all shortcuts **[Done]**
- Add find & replace feature (Ctrl+H for find, Ctrl+R for replace) **[Done]**
- Improve status and info bars for clarity (file name, mode, save status)
- Enhance UI aesthetics: borders, colors, and layout for a modern terminal look
- Update inline code comments and quick editor documentation

### Phase B: Full Editor Mode
- Implement advanced navigation, selection, and editing shortcuts
- Add multi-file/tab support (optional)
- Configurable keybindings

### Phase C: Syntax Verification
- Integrate parsers/validators for YAML, JSON, XML, INI
- Show errors inline or in a status bar

### Phase D: Polish & Ship
- Cross-platform testing (Windows, Linux, macOS)
- Finalize documentation and usage examples
- Package as binary and release

## Todos
- [x] Refactor error handling for user-friendly messages
- [x] Add unit/integration tests
- [x] Expand inline and external documentation
- [ ] Implement full editor mode with advanced shortcuts
- [ ] Add syntax verification for YAML, JSON, XML, INI
- [ ] Cross-platform validation
- [ ] Prepare release assets and documentation

## Notes
- Use Go libraries for syntax validation (e.g., go-yaml, encoding/json, encoding/xml, ini)
- Prioritize usability and stability before adding new features
- Consider user feedback for shortcut design and extensibility
