# goicy

This was forked from https://github.com/stunndard/goicy with the specific
purpose for pulling data from a different source and saving it only temporarily  
for streaming purposes

It has a slim REST API built on the golang Gin framework, and supports the

following api calls:

POST /add

When calling /add, you must supply the call with a corresponding track id
to a song from the free music archive

```
{
    track_id: string,
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
        id: int
        title: string
        artist: string
        album: string
        length: string
        url: string
        }
}
```

It takes in a single config file in json format.

## Config

```
{

}
```
