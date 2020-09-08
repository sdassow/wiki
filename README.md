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
 * `csrf-insecure bool` - send csrf cookie over http (default `false`)
 * `csrf-keyfile string` - path to csrf key file (default `./csrf.key`)
 * `data string` - data storage directory (default `./data`)
 * `git-push bool` - enable push on commit, disabled by default
 * `git-url string` - git repository to pull from (and push to), disabled by default
 * `indexdir string` - path to search index directory (default `./riot-index`)
 * `listen-address string` - address to bind to (default `:8000`)
 * `listen-network string` - network can be "tcp", "tcp4", "tcp6", "unix" or "unixpacket" (default `tcp`)
 * `listen-protocol string` - protocol can be "fcgi" or "http" (default `http`)
 * `prefix string` - URL prefix for access via a sub-path, empty by default

## Design

 - The data directory is a working copy of a git repository, and if not existent on startup, a checkout of the git repository is done
 - All wiki pages are stored inside the data directory in markdown format
 - Attachments are stored per markdown file in a directory without the mardown file extension
 - The fulltext search index is stored in a separate index directory
 - For CSRF protection a key file is used, and a random key file is generated if none exists on startup

The resulting top level directory view is:
```
./data/
      + FrontPage.md
      + FrontPage/...
./riot-index/...
./csrf.key
```

## License

MIT
