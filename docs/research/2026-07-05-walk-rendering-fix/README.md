# Pesquisa: Fix Rendering lxn/walk Dialog + GTK4 Comparação

## TL;DR

**Causa raiz do bug de rendering no walk**: O refator de layout concorrente (commit d8ada91) removeu `Layout().Update(true)` e substituiu por `Window.RequestLayout()`. O sistema de layout do dialog não completa antes de `Show()`, causando widgets invisíveis.

**GTK4 não é recomendado no Windows**: Tem problemas de rendering (bordas pretas, ignora temas), requer MSYS2, bindings imaturos (não-v1), mem leaks. walk é a melhor escolha para Windows-only.

## Fontes

- [lxn/walk#654 - Dialog children not rendering](https://github.com/lxn/walk/issues/654) - **HIGH**
- [lxn/walk#747 - WM_SIZE fix](https://github.com/lxn/walk/pull/747) - **HIGH**
- [lxn/walk#650 - Layout().Update(true) removed](https://github.com/lxn/walk/issues/650) - **HIGH**
- [gotk4 GitHub](https://github.com/diamondburned/gotk4) - **HIGH**
- [Go not suitable for desktop GUI](https://cwansart.vivaldi.net/2024/12/28/go-is-not-suitable-for-desktop-gui-apps) - **MEDIUM**

## Key Findings

### Walk Dialog Fix
1. **MaxSize height = 0** para não restringir: `MaxSize{800, 0}` (0 = sem limite)
2. **RequestLayout()** substitui `Layout().Update(true)` pós-commit d8ada91
3. **DefWindowProc** deve ser chamado em WM_WINDOWPOSCHANGED para propagar WM_SIZE/WM_MOVE
4. Dialog.Show() calcula tamanho via `MinSizeHint()` do layout. Se layout não completa, widgets ficam invisíveis.

### GTK4 no Windows
- gotk4 v0.3.1 (Aug 2024), 683 stars, **não é v1**
- GTK4 ignora temas Windows, bordas pretas feias
- Requer MSYS2 + mingw-w64-x86_64-gtk4 + gobject-introspection
- Build inicial muito lento (CGo + introspection)
- Autor tentou file manager em Go com GTK4 e desistiu, voltou para C++/GTK3

### Conclusão
**lxn/walk continua sendo a melhor escolha** para file manager Windows-only. Os bugs de rendering são conhecidos e têm workarounds.
