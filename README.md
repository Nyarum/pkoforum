# PKO Forum

A simple forum application built with Datastar and Go.

## Features

- Create new threads with title and content
- View all threads in a responsive layout
- Real-time updates using Server-Sent Events
- Clean and modern UI with Tailwind CSS

## Setup

1. Make sure you have Go installed (version 1.21 or later)

2. Install the dependencies:
```bash
go mod tidy
```

3. Run the server:
```bash
go run main.go
```

4. Open your browser and navigate to `http://localhost:8080`

## Project Structure

- `index.html` - The main frontend application
- `main.go` - The Go backend server
- `go.mod` - Go module definition and dependencies

## How it Works

The application uses:
- Datastar for frontend reactivity and backend communication
- Server-Sent Events (SSE) for real-time updates
- Gorilla Mux for routing
- In-memory storage for threads (for demonstration purposes)
- Tailwind CSS for styling 