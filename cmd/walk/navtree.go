package main

import (
	"os"
	"path/filepath"
)

func makeNavItem(text, path string, parent *NavItem) *NavItem {
	return &NavItem{text: text, path: path, parent: parent, children: nil, isSection: false}
}

func makeSection(text string) *NavItem {
	return &NavItem{text: text, path: "", parent: nil, children: nil, isSection: true}
}

func buildSectionedNavTree(model *NavTreeModel) {
	qa := makeSection("Acesso Rápido")
	for _, f := range []struct{ name, path string }{
		{"Área de Trabalho", filepath.Join(os.Getenv("USERPROFILE"), "Desktop")},
		{"Downloads", filepath.Join(os.Getenv("USERPROFILE"), "Downloads")},
		{"Documentos", filepath.Join(os.Getenv("USERPROFILE"), "Documents")},
	} {
		if _, err := os.Stat(f.path); err == nil {
			qa.children = append(qa.children, makeNavItem(f.name, f.path, qa))
		}
	}

	drives := makeNavItem("Unidades", "", nil)
	for c := 'A'; c <= 'Z'; c++ {
		drive := string(c) + ":\\"
		if _, err := os.Stat(drive); err == nil {
			drives.children = append(drives.children, makeNavItem(drive, drive, drives))
		}
	}

	thisPC := makeSection("Este PC")
	for _, f := range []struct{ name, path string }{
		{"Área de Trabalho", filepath.Join(os.Getenv("USERPROFILE"), "Desktop")},
		{"Documentos", filepath.Join(os.Getenv("USERPROFILE"), "Documents")},
		{"Downloads", filepath.Join(os.Getenv("USERPROFILE"), "Downloads")},
		{"Imagens", filepath.Join(os.Getenv("USERPROFILE"), "Pictures")},
		{"Músicas", filepath.Join(os.Getenv("USERPROFILE"), "Music")},
		{"Vídeos", filepath.Join(os.Getenv("USERPROFILE"), "Videos")},
	} {
		if _, err := os.Stat(f.path); err == nil {
			thisPC.children = append(thisPC.children, makeNavItem(f.name, f.path, thisPC))
		}
	}
	thisPC.children = append(thisPC.children, drives)
	drives.parent = thisPC

	model.roots = []*NavItem{qa, thisPC}
}

func (mw *GarqMainWindow) onNavItemSelected() {
	item := mw.navTree.CurrentItem()
	if item == nil {
		return
	}
	navItem, ok := item.(*NavItem)
	if !ok || navItem.path == "" || navItem.isSection {
		return
	}
	mw.navigateTo(navItem.path)
}

func (mw *GarqMainWindow) onNavItemActivated() { mw.onNavItemSelected() }
