#!/usr/bin/env sh

# Wait for gripmock to be ready (timeout after 20 seconds)
timeout=20
while [ $timeout -gt 0 ]; do
  if grep -q "Serving gRPC on tcp://" gripmock.log; then
    echo "gripmock is ready"
    break
  fi
  sleep 1
  timeout=$((timeout - 1))
done

if [ $timeout -eq 0 ]; then
  echo "Timeout waiting for gripmock to start"
  cat gripmock.log
  exit 1
fi 