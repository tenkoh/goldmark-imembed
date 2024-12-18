## Example: Render a remote markdown file
Think a case where a markdown file on GitHub includes some images.

```
repository/
  |- articles
  |  |- document.md
  |- images
     |- image.png
```

```sh
go build -o main main.go
# this is one of my articles about Go.
./main https://raw.githubusercontent.com/tenkoh/zenn-content/refs/heads/main/articles/range-over-func-beginner.md
```