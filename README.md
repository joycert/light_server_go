# Lights Server Go

A Go implementation of the lights server for controlling RGB lights via an ESP32 device.

## Overview

This server provides an API endpoint that accepts text messages and converts them to RGB values using OpenAI's GPT model. The RGB values are then sent to an ESP32 device to control the lights. The implementation uses Go's standard library for HTTP handling.

## Prerequisites

- Go 1.21 or later
- OpenAI API key

## Setup

1. Clone the repository:

```bash
git clone github.com/ryanjoyce/lights
cd lights/lights_server_go
```

2. Install dependencies:

```bash
go mod tidy
```

3. Create a `.env` file in the root directory with your OpenAI API key:

```
OPENAI_API_KEY=your_openai_api_key_here
```

## Running the Server

```bash
go run main.go
```

The server will start on port 8001.

## API Endpoints

### POST /messages

Send a message to control the lights.

**Request Body:**

```json
{
  "message": "Change lights to soft blue"
}
```

**Response:**

```json
{
  "status": "success",
  "message": "Message sent successfully",
  "rgb": {
    "r": 0,
    "g": 50,
    "b": 100
  }
}
```

### GET /health

Health check endpoint.

**Response:**

```json
{
  "status": "healthy"
}
```

## Configuration

- The ESP32 device URL is configured in `api/messages.go` as `ESP32_URL`.
- The OpenAI model is configured in `utils/parse_to_rgb.go` as `Model: "gpt-4o-mini"`.

## Error Handling

The server provides detailed error messages for API requests, OpenAI communication issues, and ESP32 device communication problems.
