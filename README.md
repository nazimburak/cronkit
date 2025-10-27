# cronkit

A tiny, dependency-free Go package for parsing and evaluating standard 5-field cron expressions.  
It can calculate the next matching time for a given cron expression.

---

## Features

- Supports standard 5-field cron syntax: `minute hour day month weekday`
- Handles ranges (`1-5`), steps (`*/10`), and lists (`1,5,10`)
- Fully dependency-free and lightweight
- Simple utility for parsing and computing the next run time

---

## Usage

```
cronExpr := "0 9 * * 1" // Runs at 09:00 AM every Monday.
expr, err := cronkit.Parse(cronExpr)
if err != nil {
    ...
}

nextTime := expr.Next(time.Now())
fmt.Printf("Next Scheduled Time: %s\n", nextTime)
```

## Behavior & Notes

| Behavior                      | Description                                       |
|-------------------------------|---------------------------------------------------|
| **`Day` and `Weekday` Logic** | Combined with OR semantics (*like standard cron*) |
| **Invalid Expressions**       | Cause `Parse()` to return an error                |
| **Impossible Dates**          | `Next()` returns zero (*time.Time{}*)             |
| **Infinite Loop Protection**  | Internal iteration limit prevents hanging         |
| **Time Normalization**        | Automatically handles month/day overflow          |
