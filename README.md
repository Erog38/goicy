# go-radio

This is a golang built internet radio streaming project built using 
ffmpeg bindings provided by https://github.com/giorgisio/goav .

It has a slim REST API built on the golang Gin framework, and supports the

following api calls:

POST /add

When calling /add, you must supply the call with a direct  URL
to a song in either mp3 or WAV format.

```
{
    song: string,
}
```

your response will look like:
```
{
    error: string,
    response: {
        success: boolean,
        reason: string
    }
}
```

GET /playlist

Supplies you with the current playlist information from next to last in the
queue up to 10 songs:
```
{
    error: string,
    response: {
        no_queue: boolean,
        queue: []tracks
    }
}
```

GET /current

Gives you the current track playing on the station
```
{
    error: string,
    response: {
        track: track   
    }
}
```

DELETE /remove

This uses supplied information to remove a track from the playlist
```
{
    track: track,
}
```


The response will look like the following:
```
{
    error: string,
    response: {
        success: boolean,
        reason: string,
    }
}
```

## Data Formats

Generic Response:

```
{
    error: boolean
    response: {}
}
```


### track
```
{
    track: {
        title: string
        artist: string
        album: string
        length: string
        }
}
```

It takes in a single config file in json format.

## Config

```
{

}
```
