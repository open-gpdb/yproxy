# yproxy

yproxy - a service for efficient transfer data external storages.


### debugging

1. set `debug_port` and `debug_minutes` in configuration file
2. send SIGUSR2 to `yproxy`
3. target golang tools to /debug/pprof to profile application