## 📖 Overview
This repository demonstrates the evolutionary system design of an event-driven SMS delivery pipeline. 

This project is inspired by a real-world SMS delivery system I previously worked on. I chose this specific domain because it serves as an excellent crucible for tackling the core challenges of modern backend and distributed systems. It naturally demands robust solutions for 
- high throughput
- low latency 
- fault tolerance 
- high concurrency
- tenant fairness
- strict rate limiting
- just-in-time processing
- and more 

Instead of starting with a complex distributed architecture, this project begins as a minimal, Go monolith. It progressively evolves into a decoupled system to address specific backend challenges as the hypothetical scale grows.


## 🎯 Core Service Specifications

This platform enables client businesses to reliably manage and deliver SMS communications for various use cases, such as marketing campaigns and appointment reminders. The actual physical delivery to end-user devices is handled by mobile carriers (e.g., Verizon, T-Mobile, AT&T) via their downstream REST APIs. Our system acts as the robust intermediary, ensuring millions of messages are dispatched at the requested times.

### 📍 Typical Business Scenario
1. **Schedule:** A marketer at a client company registers an SMS campaign, providing a message template, a bulk list of target users (names, phone numbers, custom variables), and a scheduled delivery time.
2. **Wait:** The system securely stores this reservation without prematurely expanding the data.
3. **Execute:** Once the scheduled time arrives, the system processes the lists, renders the individual text messages, and dispatches them to the Carrier APIs for final delivery.

### 🚧 Project Scope & System Boundary
A complete enterprise SMS platform is typically divided into two distinct parts:
1. **The Client-Facing Service (Synchronous):** A web application and REST API that handles user authentication, billing, and accepting campaign reservations.
2. **The Delivery Pipeline (Asynchronous):** The heavy-lifting backend worker group that reliably executes the scheduled jobs.

This repository omits the client-facing GUI/API. It focuses **exclusively on the asynchronous Delivery Pipeline** to highlight the engineering solutions for complex backend challenges such as high-throughput processing, external rate limiting, and fault tolerance.

## 📁 Directory Structure

Current minimal implementation structure:
- `domain/`: Contains core business models (e.g., `Campaign`).
- `worker/`: Contains the core asynchronous pipeline logic (polling the database, fetching destinations, and dispatching SMS).
- `port/`: Defines interfaces for external dependencies like Database, Storage (CSV), and Carrier API.

## 🚀 How to Run

Currently, the project is in its minimal implementation phase. The only available command is to run the test suite:

```bash
make test
```