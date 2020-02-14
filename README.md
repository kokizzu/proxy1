# tiny-http-proxy
Maybe the tiniest HTTP proxy that also has a cache.


The tiny-http-proxy acts as a reverse proxy for one server of your choice illustrated by this picture:

```
           network            proxy             cache
         .---------.       .---------.       .---------.
    <----|- - < - -|---<---|- - < - -|---<---|- < -.   |
you ---->|- - > - -|--->---|- -,- > -|--->---|- > -|   |
         |         |       |   |(*)  |       |     |   |
         |    ,-< -|---<---|< -'     |       |     |   |
         |    , ,->|--->---|- - > - -|--->---|- > -'   |
         `----+-+--´       `---------´       `---------´
              ' '
              '_'
            website

(*) When the data is not in the cache, the website will be requested and is directly stored in the cache.
```
Where "network" may be anything (LAN/WAN/...).

# Installation
Just clone this repo and run it:

```
git clone https://github.com/kokizzu/tiny-http-proxy.git
cd tiny-http-proxy
go run *.go
```

# Note

- **WARNING**: Does not work for https '__')
- Caching code from the forked repo
- Other code from [this gist](https://gist.github.com/yowu/f7dc34bd4736a65ff28d) 
