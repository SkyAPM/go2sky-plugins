# Go2sky with gin

1. [v2](v2/README.md)
1. [v3](v3/README.md)


## FAQ

### What's the difference between v2 and v3

As [commented](https://github.com/SkyAPM/go2sky/issues/59#issuecomment-645427674), v2 version will overwrite in some cases. 
To avoid this, we get the request route through `gin.context.FullPath()`
