package parser

import (
    "os"
    "fmt"
    "io"
    //"log"
    "bufio"
    "strings"
    "strconv"
)

type CronJob struct {
    Raw         string
    Schedule    []string
    User        string
    Command     string
    Description string
}

type Result struct {
    CronJobs []CronJob
}

var cronRanges = []struct {
    name string
    min int
    max int
}{
    {"minute", 0, 59},
    {"hour", 0, 23},
    {"day of month", 1,31},
    {"month", 1, 12},
    {"day of week", 0, 7},
}

func ParseCrontab(path string) (*Result, error) {
    file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	const maxCapacity = 1024*1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	//var jobs []CronJob
	jobs := make([]CronJob, 0)
	lineNo := 0

	for scanner.Scan() {
	    lineNo++
	    raw := scanner.Text()
	    trim := strings.TrimSpace(raw)
	    if trim == "" || strings.HasPrefix(trim, "#") {
	        continue
	    }
	    firstTok := strings.Fields(trim)
	    if len(firstTok) > 0 && strings.Contains(firstTok[0], "=") && !strings.HasPrefix(firstTok[0], "\"") {
	        continue
	    }

	    fields := strings.Fields(trim)
	    if len(fields) == 0 {
	        continue
	    }
	    job := CronJob { Raw: raw }

	    if strings.HasPrefix(fields[0], "@") {
	        job.Schedule = []string{fields[0]}
	        job.Description = describeSpecial(fields[0])
	        if len(fields) < 2 {
	            job.Command = ""
	            jobs = append(jobs, job)
	            continue
	        }
	        if len(fields) >= 3 {
	            job.User = fields[1]
	            job.Command = strings.Join(fields[2:], " ")
	        } else {
	            job.Command = strings.Join(fields[1:], " ")
	        }
	        jobs = append(jobs, job)
	        continue
	    }

	    if len(fields) < 6 {
	        continue
	    }

	    if err := validateSchedule(fields[:5], lineNo); err != nil {
	        fmt.Fprintf(os.Stderr, "Warning: invalid schedule on line %d", lineNo)
	        continue
	    }

	    job.Schedule = make([]string, 5)
	    copy(job.Schedule, fields[:5])
	    job.Description = DescribeSchedule(job.Schedule)
	    //if len(fields) >= 7 {
	    //    job.User = fields[5]
	    //    job.Command = strings.Join(fields[6:], " ")
	    //} else {
	    //    job.User = ""
	        job.Command = strings.Join(fields[5:], " ")
	    //}
	    jobs = append(jobs, job)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	result := &Result{}
	result.CronJobs = make([]CronJob, len(jobs))
	copy(result.CronJobs, jobs)

	return result, nil
}

func validateSchedule(fields []string, line int) error {
    for i, field := range fields {
        if err := validateField(field, cronRanges[i].min, cronRanges[i].max); err != nil {
            return fmt.Errorf("%s field '%s' invalid: %v", cronRanges[i].name, field, err)
        }
    }
    return nil
}

func validateField(field string, min, max int) error {
    if field == "*" {
        return nil
    }
    if strings.HasPrefix(field, "*/") {
        step, err := strconv.Atoi(field[2:])
        if err != nil || step <= 0 {
            return fmt.Errorf("invalid step value in '%s'", field)
        }
        return nil
    }

    parts := strings.Split(field, ",")
    for _, p := range parts {
        if strings.Contains(p, "-") {
            bounds := strings.Split(p, "-")
            if len(bounds) != 2 {
                return fmt.Errorf("invalid range '%s'", p)
            }
            start, err1 := strconv.Atoi(bounds[0])
            end, err2 := strconv.Atoi(bounds[1])
            if err1 != nil || err2 != nil {
                return fmt.Errorf("invalid range numbers in '%s'", p)
            }
            if start < min || end > max || start > end {
                return fmt.Errorf("range out of bounds (%d-%d) in '%s'", min, max, p)
            }
        } else {
            val, err := strconv.Atoi(p)
            if err != nil {
                return fmt.Errorf("invalid integer '%s'", p)
            }
            if val < min || val > max {
                return fmt.Errorf("value %d out of range (%d-%d)", val, min, max)
            }
        }
    }
    return nil
}

func describeSpecial(token string) string {
    switch token {
    case "@reboot":
        return "Run once at startup"
    case "@yearly", "@annually":
        return "Run once a year (Jan 1, 00:00)"
    case "@monthly":
        return "Run once a month (1st day, 00:00)"
    case "@weekly":
        return "Run once a week (Sunday, 00:00)"
    case "@daily", "@midnight":
        return "Run once a day (00:00)"
    case "@hourly":
        return "Run once an hour (minute 0)"
    default:
        return "Special schedule"
    }
}

func DescribeSchedule(fields []string) string {
    if (len(fields) != 5) {
        return ""
    }
    minute, hour, dom, mon, dow := fields[0], fields[1], fields[2], fields[3], fields[4]
    desc := fmt.Sprintf("At minute %s, hour %s, day-of-month %s, month %s, day-of-week %s",
        minute, hour, dom, mon, dow)
    if minute == "0" && hour == "*" && dom == "*" && mon == "*" && dow == "*" {
        return "Every hour at minute 0"
    }
    if strings.HasPrefix(minute, "*/") && hour == "*" && dom == "*" {
        return fmt.Sprintf("Every %s minutes", minute[2:])
    }
    if minute == "0" && hour == "0" && dom == "*" {
        return "Every day at midnight"
    }
    if dom == "*" && mon == "*" && dow == "1" && minute == "0" && hour == "0" {
        return "Every Monday at midnight"
    }
    return desc
}

func (result *Result) Draw(writer io.Writer) error {
    if result == nil || writer == nil {
        return nil
    }
    
    if result.CronJobs == nil {
        return nil
    }

    for _, item := range result.CronJobs {
        fmt.Fprintf(writer, "%-20s %s\n", strings.Join(item.Schedule, " "), item.Command)
    }
    return nil
}
