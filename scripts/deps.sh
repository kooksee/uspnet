#!/usr/bin/env bash

a="""



#config
github.com/spf13/cobra
github.com/spf13/viper
github.com/spf13/pflag
github.com/spf13/afero
github.com/spf13/cast
github.com/spf13/jwalterweatherman
github.com/magiconair/properties
github.com/mitchellh/mapstructure
github.com/pelletier/go-toml
gopkg.in/yaml.v2
github.com/hashicorp/hcl

#others

github.com/fsnotify/fsnotify

#test
github.com/stretchr/testify

#log
github.com/sirupsen/logrus

#net
@github.com/xtaci/kcp-go
golang.org/x/net

#wensocket
github.com/gorilla/websocket

#web
github.com/julienschmidt/httprouter

#golang
golang.org/x/crypto
golang.org/x/text

#snappy
github.com/golang/snappy

#json
github.com/json-iterator/go
github.com/v2pro/plz/reflect2

"""

for i in $a; do
    if [[ ${i:0:1} != "#" && ${i:0:1} != "@" ]];then
    gopm get -l $i
    fi
done
