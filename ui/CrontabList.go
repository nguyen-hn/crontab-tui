package ui

import (
    "github.com/jroimartin/gocui"
)

type CrontabListPanel struct {
    ViewName        string
    viewPosition    ViewPosition

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
        v.Highlight = true
        v.SelFgColor = gocui.ColorBlue
        v.SelBgColor = gocui.ColorGreen
        v.Title = " Crontab List"
    }
    return nil
}
