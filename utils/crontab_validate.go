package utils

import (
    "fmt"
    "strings"
    "os"
    "os/exec"
    "strconv"
)

func ValidateScheduleStrict(fields []string) error {
    if len(fields) != 5 {
        return fmt.Errorf("schedule must have 5 fields")
    }

    ranges := []struct{ min, max int }{
        {0, 59},  // minute
        {0, 23},  // hour
        {1, 31},  // dom
        {1, 12},  // month
        {0, 7},   // dow
    }

    for i, f := range fields {
        if err := validateField(f, ranges[i].min, ranges[i].max); err != nil {
            return fmt.Errorf("field %d: %v", i+1, err)
        }
    }
    return nil
}

func ValidateCommand(cmd string) error {
    if cmd == "" {
        return fmt.Errorf("empty command")
    }

    parts := strings.Fields(cmd)
    bin := parts[0]

    // absolute file?
    if strings.HasPrefix(bin, "/") {
        st, err := os.Stat(bin)
        if err != nil {
            return fmt.Errorf("file does not exist: %s", bin)
        }
        if st.Mode()&0111 == 0 {
            return fmt.Errorf("file is not executable: %s", bin)
        }
        return nil
    }

    // look in PATH
    if _, err := exec.LookPath(bin); err != nil {
        return fmt.Errorf("command not found in PATH: %s", bin)
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
