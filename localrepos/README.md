# localRepos

localRepos is a Key/Value map that cached the data from a particular storeDB to local memory, and keep refreshing data in certian intervals.
It's useful for small collection with high read I/O.

## interface
```go
type HashData interface {
    GetId() string
}
type LocalRepos interface {
    // ReloadEntries use to query the data and cached it to local
    ReloadEntries(ch chan HashData) error
    Close() error
}
```

