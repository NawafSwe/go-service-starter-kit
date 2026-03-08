package app

// Scheduler is implemented by every one-time job.
// Jobs are run once and exit — use them for migrations, data backfills, etc.
type Scheduler interface {
	Schedule(args ProcessArgs) error
}

// JobsMap registers one-time jobs by name.
// Example:
//
//	var JobsMap = map[string]Scheduler{
//	    "backfill-users": jobs.NewBackfillUsersJob(),
//	}
var JobsMap = map[string]Scheduler{}
