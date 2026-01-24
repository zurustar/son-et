#!/bin/bash
# Kill a process and all its children

if [ -z "$1" ]; then
    echo "Usage: $0 <PID>"
    exit 1
fi

PID=$1

# Function to kill process tree
kill_tree() {
    local pid=$1
    local children=$(pgrep -P $pid 2>/dev/null)
    
    # Kill children first
    for child in $children; do
        kill_tree $child
    done
    
    # Kill the process itself
    kill -9 $pid 2>/dev/null
}

# Kill the process tree
kill_tree $PID

echo "Process $PID and its children have been terminated"
