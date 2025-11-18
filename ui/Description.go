package ui

import (
    "github.com/jroimartin/gocui"
    "fmt"
    "crontab-tui/parser"
)

type DescriptionPanel struct {
    ViewName        string
    viewPosition    ViewPosition
}

func NewDescriptionPanel() (*DescriptionPanel, error) {
    descriptionPanel  := DescriptionPanel {
        ViewName: "description",
        viewPosition: ViewPosition{
            x0: Position{0.0, 0},
            y0: Position{0.5, 0},
            x1: Position{0.9, 1},
            y1: Position{0.9, 1},
        },
    }
    return &descriptionPanel, nil
}

func (descriptionPanel *DescriptionPanel) DrawView(g *gocui.Gui) error {
    maxX, maxY := g.Size()
    x0, y0, x1, y1 := descriptionPanel.viewPosition.GetCoordinates(maxX, maxY)
    if v, err := g.SetView(descriptionPanel.ViewName, x0, y0, x1, y1); err != nil {
        v.SelFgColor = gocui.ColorBlack
        v.SelBgColor = gocui.ColorGreen
        v.Title = " Description "
    }
    return nil
}

func (descriptionPanel *DescriptionPanel) DrawText(g *gocui.Gui, item *parser.CronJob) error {
    v, err := g.View(descriptionPanel.ViewName)
    if err != nil {
        return err
    }
    v.Clear()
    fmt.Fprintf(v, item.Description)
    return nil
}
