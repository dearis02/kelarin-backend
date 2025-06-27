#!/bin/sh

echo "Running migrations..."
./migration-tool up

echo "Initializing area..."
./init-area

exec ./server