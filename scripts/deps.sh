#!/usr/bin/env bash

a="""

#web
github.com/julienschmidt/httprouter

#config
github.com/spf13/cobra
github.com/spf13/viper
github.com/spf13/pflag
github.com/spf13/afero
github.com/spf13/cast
github.com/spf13/jwalterweatherman

github.com/fsnotify/fsnotify
github.com/gin-contrib/sse
github.com/go-kit/kit/log
github.com/go-logfmt/logfmt



#hello

"""

for i in $a; do
    if [[ ${i:0:1} != "#" ]];then
    gopm get -l $i
    fi
done
