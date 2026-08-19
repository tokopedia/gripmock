---
title: Delay Configuration
---

# Delay Configuration <VersionTag version="v3.2.16" />

Delay configuration allows you to simulate network latency and processing time in your gRPC mock responses, making your tests more realistic.

## Overview

The `delay` field in stub output configuration lets you add artificial delays before sending responses. This is useful for:
- Testing timeout scenarios
- Simulating slow network conditions
- Validating client-side retry logic
- Performance testing under realistic conditions

## Basic Usage

### JSON Format
```json
{
  "service": "TrackService",
  "method": "StreamTrack",
  "input": {
    "equals": {
      "stn": "MS#00001"
    }
  },
  "output": {
    "delay": "100ms",
    "data": {
      "stn": "MS#00001",
      "identity": "00",
      "latitude": 0.1,
      "longitude": 0.005,
      "speed": 45,
      "updatedAt": "2024-01-01T12:00:00.000Z"
    }
  }
}
```

### YAML Format
```yaml
service: TrackService
method: StreamTrack
input:
  equals:
    stn: "MS#00001"
output:
  delay: 100ms
  data:
    stn: "MS#00001"
    identity: "00"
    latitude: 0.1
    longitude: 0.005
    speed: 45
    updatedAt: "2024-01-01T12:00:00.000Z"
```

## Delay Formats

GripMock supports various time duration formats:

### Supported Units
- **Milliseconds**: `100ms`, `500ms`, `1.5ms`
- **Seconds**: `1s`, `2.5s`, `30s`
- **Minutes**: `1m`, `5m`
- **Hours**: `1h`, `2h`

### Examples
```yaml
output:
  delay: 100ms    # 100 milliseconds
  delay: 2.5s     # 2.5 seconds
  delay: 1m       # 1 minute
  delay: 1h30m    # 1 hour 30 minutes
```

## Regressive Delay

Use `delay_type: regressive` to decrease the response delay after every match. This is useful for testing clients that retry timed-out requests and eventually receive a response.

```yaml
service: ExampleService
method: GetData
input:
  contains:
    key: example
output:
  delay: 3s
  delay_type: regressive
  delay_step: 500ms
  data:
    result: success
```

The effective delay is calculated as `max(0, delay - (match_number - 1) * delay_step)`. The example produces:

```text
3s, 2.5s, 2s, 1.5s, 1s, 500ms, 0s, 0s, ...
```

- The first match uses the configured `delay`.
- The delay never becomes negative and remains at zero after reaching it.
- `delay` and `delay_step` must both be positive when the type is `regressive`.
- Omitting `delay_type`, or setting it to `default`, keeps the delay fixed.
- The match number is isolated by stub and `X-Gripmock-Session`. Without a session, it is global for the stub.
- A match advances the sequence even if the client later cancels or times out.
- Clearing or deleting the stub resets the sequence.

For unary, client-streaming, and server-streaming calls, the effective value is calculated once when the stub is matched. Bidirectional streams advance the sequence for each independently matched incoming message. A per-element `_gripmock.delay` continues to override the top-level delay.

## Streaming Responses

For streaming responses, delay is applied **before** every message in the stream.

### Uniform Delay

```yaml
service: TrackService
method: StreamTrack
input:
  equals:
    stn: "MS#00005"
output:
  delay: 200ms
  stream:
    - stn: "MS#00005"
      identity: "00"
      latitude: 0.11
      longitude: 0.006
      speed: 50
      updatedAt: "2024-01-01T13:00:00.000Z"
    - stn: "MS#00005"
      identity: "01"
      latitude: 0.11001
      longitude: 0.00601
      speed: 51
      updatedAt: "2024-01-01T13:00:01.000Z"
    - stn: "MS#00005"
      identity: "02"
      latitude: 0.11002
      longitude: 0.00602
      speed: 52
      updatedAt: "2024-01-01T13:00:02.000Z"
```

The 200ms delay is applied **before** every message:
- 200ms → Message 1 → 200ms → Message 2 → 200ms → Message 3

### Per-Element Delay <VersionTag version="v3.15.1" />

Use the reserved `_gripmock` key inside a stream element to set a per-message delay:

```yaml
output:
  stream:
    - _gripmock:
        delay: 50ms
      status: NOT_SERVING
    - _gripmock:
        delay: 150ms
      status: SERVING
    - _gripmock:
        delay: 100ms
      status: DONE
```

Each message gets its own delay instead of the global `output.delay`:

- 50ms → `{status: NOT_SERVING}`
- 150ms → `{status: SERVING}`
- 100ms → `{status: DONE}`

The `_gripmock` key is **reserved** — it does not conflict with protobuf field names
and is stripped before the message is sent to the client. The `_gripmock.delay` value
uses the same duration format as the top-level `delay` field (`100ms`, `1s`, etc.).

When a stream element contains `_gripmock.delay`, the per-element delay takes
**priority** over the global `output.delay`. Elements without `_gripmock` still
use the global delay.

This is especially useful for captured streams where `recordDelay` records
the actual inter-message timing from the upstream service. See
[Capture Mode](../modes/capture) for details.

## Use Cases

### 1. Timeout Testing
```yaml
output:
  delay: 5s
  data:
    message: "Slow response"
```
Use long delays to test client timeout handling.

### 2. Realistic Network Simulation
```yaml
output:
  delay: 150ms
  data:
    message: "Typical network latency"
```
Simulate realistic network conditions for performance testing.

### 3. Rate Limiting Simulation
```yaml
output:
  delay: 1s
  data:
    message: "Rate limited response"
```
Test client behavior under rate limiting scenarios.

## Error Responses with Delay

Delay can be combined with error responses:

```yaml
output:
  delay: 500ms
  error: "Service temporarily unavailable"
  code: 14  # UNAVAILABLE
```

## Best Practices

### 1. Realistic Delays
- Use delays that reflect real-world conditions
- Consider your application's timeout settings
- Test with various delay values

### 2. Testing Strategy
- Test with zero delay for fast responses
- Test with moderate delays (100-500ms) for normal conditions
- Test with long delays (1s+) for timeout scenarios

### 3. Performance Considerations
- Avoid extremely long delays in automated tests
- Use appropriate delays for your testing environment
- Consider the impact on test execution time

## Delay Logic

### Unary Calls
- Delay is applied **before** sending the single response
- Total delay = configured delay value

### Streaming Calls  
- Delay is applied **before** every message in the stream
- Total delay = number of messages × configured delay

### Examples

#### Unary Response (1 response)
```yaml
output:
  delay: 200ms
  data:
    message: "Hello"
```
**Timing**: 200ms delay → response sent

#### Streaming Response (3 messages)
```yaml
output:
  delay: 200ms
  stream:
    - message: "First"
    - message: "Second" 
    - message: "Third"
```
**Timing**:
- 200ms delay
- Message 1 sent
- 200ms delay
- Message 2 sent
- 200ms delay
- Message 3 sent

**Total delay**: 600ms (3 × 200ms)

## Limitations

- Maximum delay is limited by client timeout settings
- Delay affects all response types (data, error, stream)
- Delay is applied consistently across all gRPC call types

## Verification

You can verify delay behavior using gRPC clients or tools like `grpcurl`:

```bash
# Test with delay
grpcurl -plaintext -d '{"stn":"MS#00001"}' localhost:4770 TrackService/StreamTrack
```

The response time should include the configured delay plus processing time.
