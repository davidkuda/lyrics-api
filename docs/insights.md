- how to append log to a file or to a database? use a Tee on os level; Stdout and Stderr is the conventional choice.

- ListenAndServe: If you terminate the process, the last requests may get lost. Check Ardan Labs "Service" to see an alternative.
