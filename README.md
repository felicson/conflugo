CONFLUGO: help to synchronize repository documentation with Atlassian confluence
---

To build tool with credentials use the following arguments.
```bash
go build \ 
  -ldflags '-X main.ConfluenceLogin=login -X main.ConfluencePassword=password -X main.ConfluenceSpace=ME -X main.ConfluenceURL=https://example.com' \ 
  cmd/main.go
```

### How it works!

On starting conflugo check the `confluence.ancestor` file in current directory.  
File should contain the ancestor ID as a parent node for new documents tree.

If the file doesn't exist program will quiet with zero code.

This software was designed to be used inside the CI part, so we suggest that you provide configuration options during the build phase of the binary.

## Requirements
|File|Description|
|-----|----------|
|confluence.ancestor| contain  <br> ancestor ID|
|README.md | confluence page content|
| doc/*.md | directory with another documents (optional)|

[Find me on confluence](https://confluence.example.com/scope/conflugo)  
**TODO**: Add support of delete action for documents and attachments  

![](image.png)
![](https://octodex.github.com/images/yaktocat.png)