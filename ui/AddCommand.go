package ui

import (
    "github.com/jroimartin/gocui"
)

type AddCommandPanel struct {
    ViewName    string
    viewPosition ViewPosition
    HasError    bool
}

func NewAddCommandPanel() (*AddCommandPanel, error) {
    addCommandPanel := AddCommandPanel{
        ViewName: "Add Command",
        viewPosition: ViewPosition {
            x0: Position{0.1, 0},
            y0: Position{0.35, 0},
            x1: Position{0.9, 2},
            y1: Position{0.5, 2},
        },
        HasError: false,
    }
    return &addCommandPanel, nil
}

func (addCommandPanel *AddCommandPanel) DrawView( g *gocui.Gui) error {
    maxX, maxY := g.Size()
    x0, y0, x1, y1 := addCommandPanel.viewPosition.GetCoordinates(maxX, maxY)
    if v, err := g.SetView(addCommandPanel.ViewName, x0, y0, x1, y1); err != nil {
        v.SelFgColor = gocui.ColorBlack
        v.Editable = true
        v.Title = " Add crontab job "
        _, err := g.SetCurrentView(addCommandPanel.ViewName)
        if err != nil {
            return err
        }
    }
    return nil
}

