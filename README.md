# sse-eventbug-go-demo

Go backend for the ECharts demo from
[`sse-eventbus-demo`](https://github.com/ralscha/sse-eventbus-demo). The frontend
is unchanged and the backend uses
[`sse-eventbus-go`](https://github.com/ralscha/sse-eventbus-go).

Install [Task](https://taskfile.dev/) and run:

```text
task server
task client
```

Open `http://localhost:5173`. Vite proxies `/register` to the Go backend on
`http://localhost:8080`.

## License

MIT License. See [LICENSE](LICENSE) for details.
