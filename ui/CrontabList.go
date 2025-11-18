package ui

import (
    "github.com/jroimartin/gocui"
    "crontab-tui/parser"
    "fmt"
)

type CrontabListPanel struct {
    ViewName        string
    viewPosition    ViewPosition
    CrontabList     *parser.Result
}

func NewCrontabListPanel() (*CrontabListPanel, error) {
    crontabPanel := CrontabListPanel{
        ViewName: "cron",
        viewPosition: ViewPosition{
            x0: Position{0.0, 0},
            y0: Position{0.0, 0},
            x1: Position{0.9, 1},
            y1: Position{0.5, 1},
        },
        CrontabList: nil,
    }
    return &crontabPanel, nil
}

func (crontabPanel *CrontabListPanel) DrawView(g *gocui.Gui) error {
    maxX, maxY := g.Size()
    x0, y0, x1, y1 := crontabPanel.viewPosition.GetCoordinates(maxX, maxY)
    if v, err := g.SetView(crontabPanel.ViewName, x0, y0, x1, y1); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.SelFgColor = gocui.ColorBlue
        v.SelBgColor = gocui.ColorGreen
        v.Highlight = true
        if crontabPanel.CrontabList != nil {
            crontabPanel.CrontabList.Draw(v)
        }
        v.Title = " Crontab List"
    }
    return nil
}

func (crontabPanel *CrontabListPanel) DrawText(g *gocui.Gui, message string) error {
    v, err := g.View(crontabPanel.ViewName)
    if err != nil {
        return err
    }
    v.Clear()
    fmt.Fprintf(v, message)
    return nil
}
