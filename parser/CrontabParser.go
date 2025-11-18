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
    //desc := fmt.Sprintf("At minute %s, hour %s, day-of-month %s, month %s, day-of-week %s",
    //    minute, hour, dom, mon, dow)
    if strings.HasPrefix(minute, "*/") && hour == "*" && dom == "*" && mon == "*" && dow == "*" {
        return fmt.Sprintf("Every %s minutes", minute[2:])
    }
    if hour == "*" && dom == "*" && mon == "*" && dow == "*" && minute != "*" {
        return fmt.Sprintf("Every hour at :%s", minute)
    }
    if minute == "0" && hour == "0" && dom == "*" {
        return "Every day at midnight"
    }
    if dom == "*" && mon == "*" && dow == "*" && minute != "0" && hour != "*" {
        return fmt.Sprintf("Every day at %s", formatTime(hour, minute))
    }

    if dom == "1-5" && minute == "0" && hour == "0" && dom == "*" && mon == "*" {
        return "Every weekday at midnight"
    }

    if dow != "*" && dom == "*" && mon == "*" {
        return fmt.Sprintf("Every %s at %s", describeWeekdayList(dow), formatTime(hour, minute))
    }

    parts := []string{}
    
    t := describeTime(hour, minute)
    if t != "" {parts = append(parts, t)}

    if dom != "*" {
        parts = append(parts, " on the " + describeOrdinalList(dom))
    }

    if mon != "*" {
        parts = append(parts, " in " + describeMonthList(mon))
    }
    
    if dow != "*" {
        parts = append(parts, " on " + describeWeekdayList(dow))
    }

    if len(parts) == 0 {
        return "Run every minute"
    }

    return strings.Join(parts, " ")
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

func friendlyList(field string) string {
    parts := strings.Split(field, ",")
    if len(parts) == 1 {
        return parts[0]
    }

    if len(parts) == 2 {
        return parts[0] + " and " + parts[1]
    }

    return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
}

func describePart(label, field string) string {
    if field == "*" {
        if label == "minute" {
            return "every minute"
        }
        if label == "hour" {
            return " every hour"
        }
        return ""
    }

    // Step: */5 or 1-10/2
    if strings.Contains(field, "/") {
        parts := strings.Split(field, "/")
        base, step := parts[0], parts[1]

        if base == "*" {
            return fmt.Sprintf(" every %s %ss", step, label)
        }
        return fmt.Sprintf(" every %s %ss between %s", step, label, describeRange(base))
    }

    // Range: 1-5
    if strings.Contains(field, "-") {
        return fmt.Sprintf(" at each %s from %s", label, describeRange(field))
    }

    // List: 5,25,45
    if strings.Contains(field, ",") {
        return fmt.Sprintf(" at %s %ss", friendlyList(field), label)
    }

    // Single value
    return fmt.Sprintf(" at %s %s", field, label)
}

func describeRange(r string) string {
    parts := strings.Split(r, "-")
    if len(parts) != 2 {
        return r
    }
    return fmt.Sprintf("%s to %s", parts[0], parts[1])
}

func describeTime(hour, minute string) string {
    if hour == "*" && minute == "*" {
        return ""
    }

    if hour == "*" && minute != "*" {
        return fmt.Sprintf("every hour at :%s", minute)
    }

    if hour != "*" && minute == "*" {
        return fmt.Sprintf("every minute during %s", formatHour(hour))
    }

    // both fixed → HH:MM
    if hour != "*" && minute != "*" {
        return "at " + formatTime(hour, minute)
    }

    return ""
}

func formatTime(hour, minute string) string {
    h, _ := strconv.Atoi(hour)
    m, _ := strconv.Atoi(minute)

    suffix := "AM"
    if h == 0 {
        h = 12
    } else if h == 12 {
        suffix = "PM"
    } else if h > 12 {
        h -= 12
        suffix = "PM"
    }

    return fmt.Sprintf("%d:%02d %s", h, m, suffix)
}

func formatHour(hour string) string {
    return formatTime(hour, "00")
}


var monthNames = map[string]string{
    "1": "January", "2": "February", "3": "March", "4": "April",
    "5": "May", "6": "June", "7": "July", "8": "August",
    "9": "September", "10": "October", "11": "November", "12": "December",
}

var weekdayNames = map[string]string{
    "0": "Sunday", "7": "Sunday",
    "1": "Monday", "2": "Tuesday", "3": "Wednesday",
    "4": "Thursday", "5": "Friday", "6": "Saturday",
}

func describeMonthList(field string) string {
    parts := strings.Split(field, ",")

    var names []string
    for _, p := range parts {
        if name, ok := monthNames[p]; ok {
            names = append(names, name)
        } else {
            names = append(names, p)
        }
    }

    return friendlyList(strings.Join(names, ","))
}

func describeWeekdayList(field string) string {
    parts := strings.Split(field, ",")

    var names []string
    for _, p := range parts {
        if name, ok := weekdayNames[p]; ok {
            names = append(names, name)
        } else {
            names = append(names, p)
        }
    }

    return friendlyList(strings.Join(names, ","))
}

func describeOrdinalList(field string) string {
    parts := strings.Split(field, ",")

    var list []string
    for _, p := range parts {
        list = append(list, ordinal(p))
    }

    return friendlyList(strings.Join(list, ","))
}

func ordinal(s string) string {
    n, err := strconv.Atoi(s)
    if err != nil {
        return s
    }
    if n%10 == 1 && n%100 != 11 {
        return fmt.Sprintf("%dst", n)
    }
    if n%10 == 2 && n%100 != 12 {
        return fmt.Sprintf("%dnd", n)
    }
    if n%10 == 3 && n%100 != 13 {
        return fmt.Sprintf("%drd", n)
    }
    return fmt.Sprintf("%dth", n)
}
