# idempotency-go

**idempotency-go** is a small Go library that helps you make idempotent operations
easier to implement.

When building APIs or distributed systems, it’s common to face situations where
the same request might be sent multiple times (due to retries, network issues,
or client errors). Without protection, this can lead to duplicate side effects
like double charges, duplicated records, or inconsistent states.

This library provides a Wrapper around your business logic that ensures:

- The same request (with the same input) always produces the same result.
- Duplicate executions return the stored result instead of running the logic
  again.

In short: you write your business logic once, and `idempotency-go` makes sure
it runs safely, exactly once.

⚠️ **Prototype Warning:** This project is currently in the **prototype state**
and is **not ready for production use**.

## Action, Input, and Outputs

At the heart of `idempotency-go` is the [Action](https://github.com/ymz-ncnk/idempotency-go/blob/main/action.go)
— your business logic function.

```go
type Action[T, I, S any] func(ctx context.Context, repos T, input I) (S, error)
```

- Repositories (T): A bundle of data-access dependencies, provided by the
  `UnitOfWork`. This is where your Action reads and writes data (e.g., database
  tables, domain repositories). One of these repositories must always be the
  idempotency store.
- Input (I): The parameters passed to your `Action` (like service DTO). The
  input type must implement the `Hasher` interface so it can be uniquely
  identified and checked for consistency.
- Success Output (S): The normal result your `Action` produces when everything
  goes well.
- Failure Output: A structured form of a business failure (e.g., “insufficient
  funds” or “stock unavailable”). This is not returned directly. Instead, it’s
  mapped to an error so callers only ever see success or error.

After the `Action` is executed, its result (whether success or failure) is
persisted in the idempotency store. On future retries with the same input, the
stored output is returned immediately without re-running the `Action`.

Both the `Action` execution and the storage of its result are handled inside a
single `UnitOfWork`. This ensures atomicity: either both succeed, or both are
rolled back.

## Example

Let’s say we want to build a money transfer API that should not execute the
same transfer twice. With `idempotency-go`, we wrap our business logic in an
`Action` and let the library handle idempotency for us.

```go
type TransferInput struct { 
  FromAccount string 
  ToAccount string 
  Amount int64 
}

// Implement the Hasher interface so the input can be uniquely identified.
func (in TransferInput) Hash() (string, error) {
  return fmt.Sprintf("%s:%s:%d", in.FromAccount, in.ToAccount, in.Amount), nil 
}

// Define the repositories available inside the UnitOfWork.
type RepositoryBundle struct {
  accountsRepo AccountsRepo 
  store idempotency.Store 
}

func (r RepositoryBundle) IdempotencyStore() idempotency.Store { 
  return r.store 
}

// Define the success output. 
type TransferResult struct { 
  TransactionID string 
}

// Define the failure output (structured business error). 
type TransferFailure struct { 
  Reason string 
}
```

Now we can set up the wrapper:

```go
var (
  unitOfWork = uow.NewMemDBUnitOfWork(db, factory) // Only in memory
  // implementation available at the moment. The factory ensures a new
  // repository bundle is created for each transaction.

  // The Manager is responsible for storing and retrieving both success and
  // failure outputs. It uses serializers to persist them, and a converter to
  // turn stored failures back into errors.
  manager = idempotency.NewManager( 
    serializer.JSONSerializer[TransferResult]{},
    serializer.JSONSerializer[TransferFailure]{}, 
    // Converts a stored failure output back into an error.
    func(f TransferFailure) error { return ErrInsufficientFunds }, 
  )
)
wrapper := idempotency.NewWrapper(unitOfWork, manager, 
  // This function determines which Go errors should be persisted as a failure
  // output (return 'true').
  func(err error) (TransferFailure, bool) { 
    if errors.Is(err, ErrInsufficientFunds) { 
      return TransferFailure{Reason: err.Error()}, true 
    }
    // Other errors (like lost DB connection) will not be persisted.
    return TransferFailure{}, false 
  }, 
)
```

And finally, wrap the Action:

```go
transferAction := func(ctx context.Context, repos RepositoryBundle, 
  input TransferInput) (TransferResult, error) { 
  ...
}
var (
  idempotencyKey = "transfer-123"
  input = TransferInput{ 
    FromAccount: "A", 
    ToAccount: "B", 
    Amount: 100, 
  }
)
result, err := wrapper.Wrap(ctx, idempotencyKey, input, transferAction)
```

- The first call runs the transfer and stores the result.
- A repeated call with the same input and ID (transfer-123) returns the stored
  result immediately.
- A repeated call with the same ID but different input fails with a hash
  mismatch error.

A complete, working example illustrating the full component setup can be found
in the [integration_test package](https://github.com/ymz-ncnk/idempotency-go/tree/main/integration_test).
