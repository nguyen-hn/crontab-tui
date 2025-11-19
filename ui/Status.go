package ui

import (
    "github.com/jroimartin/gocui"
    "fmt"
)

type StatusPanel struct {
    ViewName string
    viewPosition ViewPosition
}

func NewStatusPanel() (*StatusPanel, error) {
    statusPanel := StatusPanel{
        ViewName: "status",
        viewPosition: ViewPosition {
            x0: Position{0.0, 0},
            y0: Position{0.95, 1},
            x1: Position{1.0, 1},
            y1: Position{1.0, 1},
        },
    }
    return &statusPanel, nil
}

func (statusPanel *StatusPanel) DrawView(g *gocui.Gui) error {
    maxX, maxY := g.Size()
    x0, y0, x1, y1 := statusPanel.viewPosition.GetCoordinates(maxX, maxY)
    if v, err := g.SetView(statusPanel.ViewName, x0, y0, x1, y1); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.SelFgColor = gocui.ColorBlue
        v.SelBgColor = gocui.ColorGreen
        v.Frame = false
        fmt.Fprintln(v, "j: Down\tk: Up")
    }
    return nil
}
