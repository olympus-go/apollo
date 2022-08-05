![apollo](./apollo.png)
# apollo

This library provides bindings around `librespot-golang` to help streamline creating a logical player and queueing tracks
to be played.

### Create a player
```go
player := spotify.NewPlayer()
if err := player.Login(); err != nil {
	panic(err)
}
```

### Start streaming tracks
The player supports searching and streaming track data.
```go
trackResults, err := player.SearchTrack("your graduation", 1)
if err != nil {
	panic(err)
}

player.QueueTrack(trackResults[0])
```

### Full example
```go
player := spotify.NewPlayer()
if err := player.Login(); err != nil {
    panic(err)
}
player.EnableAutoplay()

trackResults, err := player.SearchTracks("your graduation", "", "", 10)
player.QueueTrack(trackResults[0])

```

### Register an encoding function
