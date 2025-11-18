package main

import (
    "log"
    "github.com/jroimartin/gocui"
    "crontab-tui/ui"
    "fmt"
    "os"
    //"strings"
    "crontab-tui/parser"
)

var crontablistPanel *ui.CrontabListPanel
var descriptionPanel *ui.DescriptionPanel
var cursor *ui.Cursor

func main() {
    g, err := gocui.NewGui(gocui.OutputNormal)
    if err != nil {
        log.Panicln(err)
    }
    defer g.Close()

    g.SetManagerFunc(layout)
    crontablistPanel, _ = ui.NewCrontabListPanel()
    descriptionPanel, _ = ui.NewDescriptionPanel()
    cursor = &ui.Cursor{}
    
    const path = "./example.txt"
    jobs, err := parser.ParseCrontab(path)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error %v\n", err)
        os.Exit(1)
    }

    //fmt.Printf("Parsed %d job(s) from %s\n\n", len(jobs.CronJobs), path)

    crontablistPanel.CrontabList = jobs
    crontablistPanel.DrawView(g)
    descriptionPanel.DrawView(g)
    descriptionPanel.DrawText(g, &crontablistPanel.CrontabList.CronJobs[0])

    g.SetCurrentView(crontablistPanel.ViewName)

    keybindings(g)
    
    if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
        log.Panicln(err)
    }


	//for _, j := range jobs.CronJobs {
	//	//fmt.Printf("Line %d: raw: %s\n", j.LineNumber, j.Raw)
	//	// Show schedule parts and lengths
	//	if len(j.Schedule) == 1 && strings.HasPrefix(j.Schedule[0], "@") {
	//		fmt.Printf("  Schedule special: %s (len=%d)\n", j.Schedule[0], len(j.Schedule[0]))
	//	} else {
	//		for i, p := range j.Schedule {
	//			fmt.Printf("  Schedule part %d: '%s' (len=%d)\n", i+1, p, len(p))
	//		}
	//	}
	//	if j.User != "" {
	//		fmt.Printf("  User: '%s' (len=%d)\n", j.User, len(j.User))
	//	} else {
	//		fmt.Printf("  User: <none>\n")
	//	}
	//	fmt.Printf("  Command: '%s' (len=%d)\n", j.Command, len(j.Command))
	//	fmt.Printf("  Description '%s' \n\n", j.Description)
	//}
}

func layout(g *gocui.Gui) error {
    render(g)
    return nil
}

func render(g *gocui.Gui) {
    crontablistPanel.DrawView(g)
    descriptionPanel.DrawView(g)
}

func keybindings(g *gocui.Gui) {
    if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(crontablistPanel.ViewName, 'q', gocui.ModNone, quit); err != nil {
	    log.Panicln(err)
	}
	if err := g.SetKeybinding(crontablistPanel.ViewName, 'k', gocui.ModNone, cursorMovement(-1)); err != nil {
	    log.Panicln(err)
	}
	if err := g.SetKeybinding(crontablistPanel.ViewName, 'j', gocui.ModNone, cursorMovement(1)); err != nil {
	    log.Panicln(err)
	}
}

func exit(g *gocui.Gui, v *gocui.View) error {
    return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
    return gocui.ErrQuit
}

func cursorMovement(d int) func(g *gocui.Gui, v *gocui.View) error {
    return func(g *gocui.Gui, v *gocui.View) error {
        cursor.Move(g, v, d, func(yOffset int, yCurrent int) error {
            if g.CurrentView().Name() == crontablistPanel.ViewName {
                if yOffset + yCurrent >= len(crontablistPanel.CrontabList.CronJobs) {
                    return nil
                }
                descriptionPanel.DrawText(g, &crontablistPanel.CrontabList.CronJobs[yOffset+yCurrent])
            }
            return nil
        })
        return nil
    }
}
