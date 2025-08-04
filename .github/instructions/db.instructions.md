---
applyTo: "**"
---

# Database Usage Instructions for maya-canteen

## Overview
- The project uses a SQLite database by default, located at `db/canteen.db`.
- The database path is configurable via the `BLUEPRINT_DB_URL` environment variable.
- The database path is set at application startup in `cmd/api/main.go` using `server.SetupDBPath()`.
- If `BLUEPRINT_DB_URL` is not set, the path defaults to a `db/canteen.db` file in the executable's directory.

## Key Files
- `internal/server/db_path.go`: Contains `SetupDBPath()` which determines the DB path.
- `cmd/api/main.go`: Sets the environment variable and initializes the DB path at startup.

## How it Works
1. On startup, `main.go` calls `server.SetupDBPath()` to determine the DB path.
2. The result is set as the `BLUEPRINT_DB_URL` environment variable.
3. All DB access in the app uses this environment variable for the DB location.

## Customizing the DB Path
- To use a custom DB file, set the `BLUEPRINT_DB_URL` environment variable before running the app:
  ```sh
  export BLUEPRINT_DB_URL=/path/to/your.db
  make run
  ```
- If not set, the app will use the default path relative to the executable.

## Example
- Default: `db/canteen.db` (created if not present)
- Custom: `export BLUEPRINT_DB_URL=/tmp/test.db`

## Troubleshooting
- If the DB file is missing or the path is invalid, errors will be logged at startup.
- Check logs for messages from `SetupDBPath()` if DB issues occur.

---
For more, see `internal/server/db_path.go` and `cmd/api/main.go`.
