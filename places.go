package main

import (
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func (a *app) places(c *ui.Context, t *ui.Theme) {
	ui.Column(c).Padding(8, 10).Gap(6).Radius(8).Background(t.Surface).Border(1, t.Border).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
			ui.Text(c, "Custom directories").Bold().Grow(1)
			a.iconButton(c, iconPlus, "Add directory", "Add a directory", true, a.beginAdd)
			pending := a.drafting || a.confirmDel != ""
			tip := "Save the directory"
			if a.confirmDel != "" {
				tip = "Remove the directory"
			}
			a.iconButton(c, iconCheck, "Confirm", tip, pending, func() { _ = a.confirmPending() })
			a.iconButton(c, iconX, "Cancel", "Cancel", pending, a.cancelPending)
		})
		if a.drafting {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.TextInput(c, &a.draftName).Label("Directory name").Placeholder("Name").Width(180)
				ui.TextInput(c, &a.draftPath).Label("Directory path").Placeholder("Path").Grow(1)
				a.iconButton(c, iconFolder, "Choose folder", "Choose a folder", true, a.browseDraft)
			})
		}
		if len(a.customs) == 0 && !a.drafting {
			ui.Text(c, "None yet.").TextColor(t.TextMuted)
		}
		if len(a.customs) > 0 {
			ui.Scroll(c).MaxHeight(168).Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					for _, d := range a.customs {
						a.customRow(c, t, d)
					}
				})
			})
		}
	})
}

func (a *app) customRow(c *ui.Context, t *ui.Theme, d CustomDir) {
	id := d.ID
	row := ui.Row(c).Padding(4, 6).Gap(8).AlignItems(ui.Center).Radius(6)
	if a.confirmDel == id {
		row = row.Background(ui.Hex("#d93a35").Alpha(0.12))
	}
	row.Children(func() {
		ui.Column(c).Grow(1).Gap(1).Children(func() {
			ui.Text(c, d.Name).SingleLine()
			ui.Text(c, tilde(a.home, d.Path)).SingleLine().FontSize(12).TextColor(t.TextMuted)
		})
		if a.confirmDel == id {
			ui.Text(c, "Remove?").TextColor(t.TextMuted).Shrink(0)
		}
		a.iconButton(c, iconDuplicate, a.opLabel("Duplicate", d), "Duplicate", a.confirmDel == "", func() {
			a.beginDuplicate(id)
		})
		a.iconButton(c, iconTrash, a.opLabel("Delete", d), "Delete", true, func() {
			a.askDelete(id)
		})
	})
}

func (a *app) browseDraft() {
	if a.pickDir != nil {
		p, err := a.pickDir()
		if err != nil || strings.TrimSpace(p) == "" {
			return
		}
		a.applyPicked(p)
		return
	}
	win := a.win
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:      win,
			Title:       "Choose a directory",
			Directory:   true,
			ButtonLabel: "Choose",
		})
		if err != nil || len(paths) == 0 || win == nil {
			return
		}
		picked := paths[0]
		win.Update(func() {
			a.applyPicked(picked)
		})
	}()
}

func (a *app) applyPicked(p string) {
	a.drafting = true
	a.placesOpen = true
	a.draftPath = p
	if strings.TrimSpace(a.draftName) == "" {
		a.draftName = filepath.Base(p)
	}
}
