# wiking

Golang based wiki engine with content in Markdown format.

Additional features:

 - Git support using [go-git](https://github.com/go-git/go-git) (pure Go implementation)
 - Attachments
 - Diagrams using [Mermaid](https://mermaid-js.github.io/mermaid/)
 - Fulltext search using [Riot](https://github.com/go-ego/riot) engine

## Configuration

By default the web server listens on localhost port 8000, and wiki pages are stored in `./data`.
This can be changed with environment variables or a configuration file.

The configuration options are as follows:

 * `brand string` - branding at top of each page (default `Wiki`)
 * `cookie-insecure bool` - send cookies over http (default `false`)
 * `cookie-keyfile string` - path to cookie keyfile (default `./cookie.key`)
 * `data string` - data storage directory (default `./data`)
 * `git-push bool` - enable push on commit, disabled by default
 * `git-url string` - git repository to pull from (and push to), disabled by default
 * `hosts list-of-strings` - hostnames to allow requests to, protecting against dns rebind attacks, and used for dynamic TLS certificate when protocol is "https" and no certificate and keyfile was provided, defaults to `localhost` and `127.0.0.1`
 * `indexdir string` - path to search index directory (default `./riot-index`)
 * `listen-address string` - address to bind to (default `:8000`)
 * `listen-network string` - network can be "tcp", "tcp4", "tcp6", "unix" or "unixpacket" (default `tcp`)
 * `listen-protocol string` - protocol can be "fcgi", "http", or "https" (default `http`)
 * `prefix string` - URL prefix for access via a sub-path, empty by default
 * `tls-certfile string` - path to TLS certificate file in PEM format, empty by default
 * `tls-keyfile string` - path to TLS key file in PEM format, empty by default

## Design

 - The data directory is a working copy of a git repository, and if not existent on startup, a checkout of the git repository is done
 - All wiki pages are stored inside the data directory in markdown format
 - Attachments are stored per markdown file in a directory without the mardown file extension
 - The fulltext search index is stored in a separate index directory
 - For CSRF and cookie protection a key file is used, and a random key file is generated if none exists on startup

The resulting top level directory view is:
```
./data/
      + FrontPage.md
      + FrontPage/...
./riot-index/...
./cookie.key
```

## License

MIT
