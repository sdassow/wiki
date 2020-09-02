# wiking

Golang based wiki engine with content in Markdown format.

Additional features:

 - Git support using [go-git](https://github.com/go-git/go-git) (pure Go implementation)
 - Attachments
 - Diagrams using [Mermaid](https://mermaid-js.github.io/mermaid/)
 - Fulltext search using [Riot](https://github.com/go-ego/riot) engine

## Configuration

By default the web server listens on localhost port 8000, and wiki pages are stored in `./data`.
This can be changed with command line options or a configuration file.

The command line options are as follows:

 * `--bind [host]:port` - address to bind to (default `0.0.0.0:8000`)
 * `--brand string` - branding at top of each page (default `Wiki`)
 * `--config file` - configuration file
 * `--data dir` - data storage directory (default `./data`)
 * `--git-url url` - git repository to pull from (and push to), disabled by default
 * `--git-push` - enable push on commit, disabled by default

## License

MIT
