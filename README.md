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

 * `bind [host]:port` - address to bind to (default `0.0.0.0:8000`)
 * `brand string` - branding at top of each page (default `Wiki`)
 * `csrf-insecure bool` - send csrf cookie over http (default `false`)
 * `csrf-keyfile string` - path to csrf key file (default `./csrf.key`)
 * `data dir` - data storage directory (default `./data`)
 * `git-push` - enable push on commit, disabled by default
 * `git-url url` - git repository to pull from (and push to), disabled by default
 * `indexdir string` - path to search index directory (default `./riot-index`)

## License

MIT
