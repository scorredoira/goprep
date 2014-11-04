goprep
========

A tool to comment or uncomment code based on #if and #endif 
directives.

For example:

```Go
       //#if debug
       fmt.Println("received", a, b)
       //#endif
```

```bash
 $ goprep
```
will comment the line:
```Go
       //#if debug
       //fmt.Println("received", a, b)
       //#endif
```

```bash
 $ goprep debug
```
will uncomment it back:

 It calls [goimports](https://github.com/bradfitz/goimports) to update imported packages that may be
 needed by the uncommented code.