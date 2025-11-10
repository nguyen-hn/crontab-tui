package main

import (
    "log"
    "github.com/jroimartin/gocui"
    "crontab-tui/ui"
)

var crontablistPanel *ui.CrontabListPanel
var descriptopnPanel *ui.DescriptionPanel

func main() {
    g, err := gocui.NewGui(gocui.OutputNormal)
    if err != nil {
        log.Panicln(err)
    }
    defer g.Close()

    g.SetManagerFunc(layout)
    crontablistPanel, _ = ui.NewCrontabListPanel()
    descriptopnPanel, _ = ui.NewDescriptionPanel()
    
    crontablistPanel.DrawView(g)
    descriptopnPanel.DrawView(g)

    keybindings(g)
    
    if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
        log.Panicln(err)
    }
}

func layout(g *gocui.Gui) error {
    render(g)
    return nil
}

func render(g *gocui.Gui) {
    crontablistPanel.DrawView(g)
    descriptopnPanel.DrawView(g)
}

func keybindings(g *gocui.Gui) {
    if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
    return gocui.ErrQuit
}
