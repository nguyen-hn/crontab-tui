package main

import (
    "log"
    "github.com/jroimartin/gocui"
    "crontab-tui/ui"
    "fmt"
    "os"
    "time"
    "strings"
    "crontab-tui/parser"
    "crontab-tui/utils"
)

var crontablistPanel *ui.CrontabListPanel
var descriptionPanel *ui.DescriptionPanel
var addCommandPanel  *ui.AddCommandPanel
var statusPanel      *ui.StatusPanel
var cursor *ui.Cursor
var CRON_FILE = "./example.txt"

func main() {
    g, err := gocui.NewGui(gocui.OutputNormal)
    if err != nil {
        log.Panicln(err)
    }
    defer g.Close()

    g.SetManagerFunc(layout)
    crontablistPanel, _ = ui.NewCrontabListPanel()
    descriptionPanel, _ = ui.NewDescriptionPanel()
    addCommandPanel,  _ = ui.NewAddCommandPanel()
    statusPanel, _      = ui.NewStatusPanel()
    cursor = &ui.Cursor{}
    
    //const path = "./example.txt"
    jobs, err := parser.ParseCrontab(CRON_FILE)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error %v\n", err)
        os.Exit(1)
    }

    crontablistPanel.CrontabList = jobs
    crontablistPanel.DrawView(g)
    descriptionPanel.DrawView(g)
    statusPanel.DrawView(g)
    descriptionPanel.DrawText(g, &crontablistPanel.CrontabList.CronJobs[0])
    //fmt.Printf("Parsed %d job(s) from %s\n\n", len(jobs.CronJobs), path)

    parser.WatchCronFile(CRON_FILE, time.Second, func() {
        g.Update(func(gui *gocui.Gui) error {
            new_jobs, err := parser.ParseCrontab(CRON_FILE)
            if err != nil {
                return err
            }
            crontablistPanel.CrontabList = new_jobs
            crontablistPanel.DrawView(gui)
            descriptionPanel.DrawView(gui)
            descriptionPanel.DrawText(gui, &crontablistPanel.CrontabList.CronJobs[0])
            return nil
        }) 
    })


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
	if err := g.SetKeybinding("", gocui.KeyCtrlF, gocui.ModNone, drawAddEditor); err != nil {
	    log.Panicln(err)
	}
	if err := g.SetKeybinding(addCommandPanel.ViewName, gocui.KeyEnter, gocui.ModNone, addCrontabJob); err != nil {
	    log.Panicln(err)
	}
    if err := g.SetKeybinding(addCommandPanel.ViewName, gocui.KeyEsc, gocui.ModNone, clearErrorOnType); err != nil {
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

func drawAddEditor(g *gocui.Gui, _ *gocui.View) error {
    err := addCommandPanel.DrawView(g)
    if err != nil {
        return err
    }
    return nil
}

func addCrontabJob(g *gocui.Gui, v *gocui.View) error {
    input := strings.TrimSpace(v.Buffer())
    if input == "" {
        g.DeleteView(addCommandPanel.ViewName)
        g.SetCurrentView(crontablistPanel.ViewName)
        return nil
    }
    fields := strings.Fields(input)
    if len(fields) < 6 {
        return redrawPopupError(g, v, "Invalid format.\nUse: M H DOM MON DOW COMMAND")
    }
    schedule := fields[:5]
    command := strings.Join(fields[5:], " ")
    if err := utils.ValidateScheduleStrict(schedule); err != nil {
        return redrawPopupError(g, v, "Schedule error:\n"+err.Error())
    }

    if err := utils.ValidateCommand(command); err != nil {
        return redrawPopupError(g, v, "Command error:\n"+err.Error())
    }

    if err := AppendCrontabJob(CRON_FILE, schedule, command); err != nil {
        return redrawPopupError(g, v, "Cannot write:\n"+err.Error())
    }
    

    g.DeleteView(addCommandPanel.ViewName)
    g.SetCurrentView(crontablistPanel.ViewName)
    return nil
}

func redrawPopupError(g *gocui.Gui, v *gocui.View, msg string) error {
    g.Update(func(g *gocui.Gui) error {
        v.Clear()
        // Red text (ANSI escape)
        fmt.Fprintf(v, "\033[31m%s\033[0m\n\n", msg)
        fmt.Fprintf(v, "Please fix and press ENTER:\n")
        fmt.Fprintln(v, "")
        x, y := v.Cursor()
        v.SetCursor(x, y)
        v.SetOrigin(0, y)
        return nil
    })
    return nil
}

func AppendCrontabJob(filePath string, schedule []string, command string) error {
    f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()
    line := fmt.Sprintf("%s %s\n", strings.Join(schedule, " "), command)
    _, err = f.WriteString(line)
    return err
}

func clearErrorOnType(g *gocui.Gui, v *gocui.View) error {
    if !addCommandPanel.HasError {
        return nil // nothing to do
    }

    addCommandPanel.HasError = false

    g.Update(func(gui *gocui.Gui) error {
        v.Clear()
        v.SetCursor(0, 0)
        v.SetOrigin(0, 0)
        return nil
    })

    return nil
}

func closePopUp(g *gocui.Gui, v *gocui.View) error {
    return nil
}
