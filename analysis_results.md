# N+1 Query Analysis Results

I've analyzed the core loops across your `internal/services/` and `engine/` directories. An "N+1 query problem" occurs when your code makes a database query inside a loop, meaning if there are $N$ items in the loop, you make $N$ separate trips to the database instead of 1.

Here are the specific locations where N+1 loops are currently executing in your project:

## 1. `subscriberService.Identify` (Dynamic N+1)
**File:** `internal/services/subscriber.go` (Line 64)

When a user calls your API to identify a subscriber, they can pass multiple `Contacts` (e.g., an email and a phone number). You loop over these contacts and call `UpsertSubscriber` for each one:

```go
// If input.Contacts has 5 items, this loop makes 5 separate DB queries sequentially
for _, contact := range input.Contacts {
    // ...
    subscriber, err := s.repo.UpsertSubscriber(ctx, params)
}
```

## 2. `Engine.ingestSystem` (High-Impact N+1)
**File:** `engine/notification/core/engine.go` (Line 92)

When a system event triggers (e.g., a billing limit alert), the engine fetches all owners of the workspace. It then loops over the owners and processes the notification for each one sequentially. 

```go
for _, contact := range owners {
    // ingestNormal makes ~4 DB calls internally.
    // If a workspace has 10 owners, this loop makes 40 sequential DB queries!
    if err := e.ingestNormal(ctx, ic); err != nil {
        continue
    }
}
```

## 3. `Engine.ingestNormal` (High-Impact N+1)
**File:** `engine/notification/core/engine.go` (Line 124)

When a standard notification is triggered, the engine resolves which channels to send to (Email, SMS, Push, etc.). It loops over these channels sequentially:

```go
for _, ch := range channels {
    // ingestChannel makes DB calls for billing limits, idempotency checks, 
    // opt-out preferences, and logging. 
    // Sending to 3 channels = 12 sequential DB queries.
    if err := e.ingestChannel(ctx, ic); err != nil {
        continue
    }
}
```

## 4. `workspaceService.setupNewWorkspace` (Fixed N+1)
**File:** `internal/services/workspace.go` (Line 154)

When a new workspace is created, you loop over an array to create the default environments. Because this array is hardcoded to 2 (`development` and `production`), it only ever makes 2 queries. It's technically an N+1, but the performance impact is negligible.

```go
defaultEnvs := []string{"development", "production"}
for _, val := range defaultEnvs {
    err := repo.CreateEnvironment(ctx, ...)
}
```

---

> [!TIP]
> **How to solve dynamic N+1s?**
> Because the size of `contacts`, `owners`, and `channels` is dynamic, you cannot use your `parallel.Query2` helper. The best solution is to add a new generic `RunBatch` function to your `pkg/parallel` package that takes an array of functions and runs them concurrently using `errgroup`.
