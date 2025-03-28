# TTK4145 Real-Time programming project group 67

This project is developed as part of the TTK4145 Real-Time Programming course at NTNU. The project uses master-slave logic to assign orders. Every elevator has knowledge of all orders, through peer-to-peer communication logic. Ensuring that every elevator alwyas know each others orders and networkstatus, resulting in a smooth takeover if the master dies. 

## How to run the program:
Run the program using: "go run main.go -id=... -port=...". The id and port has to be a number. 
Setup the elevatorserver in the terminal: "elevatorserver --port ...". 
The ports have to be the same number. Each elevator must have a unique port.

### elevatorDriver:
Hardware interface. Polls elevator inputs and it controls outputs. Also contains the elevator FSM and utility functions for movement and direction.

### orders:
Handles button presses and local/global order matrices. Controls light updates for cab and hall buttons.

### master:
Runs the centralized assignment logic (when elected as master), based on the state of all elevators. Reassigns orders on stateupdate, new orderEvent and state failure. Uses hall_request_assigner executable to compute the optimal order sequence.

### network:
Manages network communication using UDP broadcast, ensuring communication between the elevators. Includes peer detection, masterselection logic, peerconnection monitoring.

### types:
Contains shared constants and type definitions used across modules (e.g., elevator state, directions, button events).

### main.go
Entry point that initializes all modules and starts concurrent goroutines for networking, FSM, master and order handling.

