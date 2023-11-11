#!/bin/bash
xvfb=$(pgrep Xvfb)
if [ -z "$xvfb" ]; then
  rm -f /tmp/.X10-lock
  Xvfb :10 -ac -screen 0 1024x768x8 &
  echo "Started Xvfb in pid $!"
else
  echo "Xvfb already running"
fi