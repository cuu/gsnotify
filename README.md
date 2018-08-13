A simple notify widget for  [GameShell](https://www.clockworkpi.com/)  

In this folder:  
> ~/apps/launcher/sys.py/gsnotify/Jobs

you can create any script/program with return data in json format.  

The returned json format is as follows:  

For notify once:  
> {"type":"once","content":"Hi! I am the anti-addiction robot."}

For repeated notice:  
> {"type":"repeat","content":"Have you done your homework yet?"}  

The script file extension supports by default: **".sh",".py",".lsp",".js",".bin"**    
Note that your script has executable permissions. (chmod +x yourscript.sh)  

Here are 2 examples for bash scripts  

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

The notify widget configuration file named "gsnotify.cfg" in this folder:

> ~/apps/launcher/sys.py/gsnotify  

And the meaning of each parameter as follows:

* DELAY_FREQ for polling interval, the default value is 30000, which means 30 seconds.
* BGCOLOR for background color, the default value is #eab934
* TXTCOLOR for font color, the default value is #ffffff
* FTSIZE for font size, the default value is 14(px).
* Width for notify widget width, the default value is 320(px).
* Height for notify widget height, the default value is 20(px).
