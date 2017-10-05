#!/bin/sh

make && git commit -am "`date`" && git pull && git push && echo "cd /home/pi/go/src/github.com/iuyte/xkcd/ && git pull && git push && TZ=US/Eastern sudo /home/pi/xkcd.service reload &" | ssh pi@raspberrypi.local
