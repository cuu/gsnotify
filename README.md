Simple notify widget for GameShell

Running scripts Under Jobs every 30 seconds,and display the results on top of screen 

The script will receive a parameter of unixstamp time ( in bash , it's $1)


The script returns a json formatted data that interacts with the gsnotify

Format is
{"type":"once","content":"the is the content "}
Once type is only shown once

{"type":"repeat","content":"the is the content "}
The repeat type will be repeated all the time as long as it meets the criteria

Very simple

For now it supports file with ".sh",".py",".lsp",".js",".bin"  exts
and the script file must have have executable permissions (chmod +x whatever.sh)


Here is two example bash scripts:

01\_test.sh

```
#!/bin/bash
echo '{"type":"repeat","content":"this is a test"}'
```

02\_time.sh, display time info every 60 seconds

```
#!/bin/bash

SLICE=60 ## 60 seconds

TIME_FILE="02_time.txt"

if [ ! -f $TIME_FILE ]; then
	echo "No time record"
	echo $1 > $TIME_FILE
fi


TIME1=`cat $TIME_FILE`
TIME2=$1

RES=`expr $TIME2 - $TIME1`


if [ $RES -gt $SLICE ]; then
	timestr=`date -d @$TIME2 +%D%T`
    echo $TIME2 > $TIME_FILE
	echo "{\"type\":\"once\",\"content\":\"$timestr\"}"
fi
```



