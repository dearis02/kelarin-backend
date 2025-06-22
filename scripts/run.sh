#!/bin/sh

echo "Running migrations..."
./migration-tool up

exec ./server