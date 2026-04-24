# Network flow

### Join Game
```mermaid
sequenceDiagram
    
    participant C1 as Client1
    participant S as Server
    participant C2 as Client2


    note right of C1: Client1 Clicks Login
    C1->>S: 1. Login/Register Request (TCP)
    S-->>C1: 2. Login Response (TCP) [Session]
    note right of C1: Client1 Clicks Match
    C1->>S: 3. OpJoin [Session]
    note over S: Room State: Waiting
    S-->>C1: 4. OpJoinAck [Assigned PlayerID]
    S-->>C1: 5. OpRoomUpdate [all PlayerIDs, RoomState]
    note right of C1: Client1 Enters Waiting
    note left of C2: Client2 Clicks Login
    C2->>S: 1. Login/Register Request (TCP)
    S-->>C2: 2. Login Response (TCP) [Session]
    note left of C2: Client2 Clicks Match
    C2->>S: 3. OpJoin [Session]
    note over S: Room State: Ready
    S-->>C2: 4. OpJoinAck [Assigned PlayerID]
    C2<<-->>C1: 5. OpRoomUpdate [all PlayerIDs, all ReadyStates]
    note right of C1: Client1 Enters Ready
    note left of C2: Client2 Enters Ready

    note right of C1: Client1 Clicks Ready
    C1->>S: 5. OpReady [PlayerID]
    note over S: Player1 Ready
    S-->>C1: 7. OpRoomUpdate [all PlayerIDs, all ReadyStates]
    note left of C2: Client2 Clicks Ready
    C2->>S: 5. OpReady [PlayerID]
    note over S: Player2 Ready
    C1<<-->>C2: 7. OpRoomUpdate [all PlayerIDs, all ReadyStates]
    note right of C1: Client1 Enters Playing
    note left of C2: Client2 Enters Playing
```

### LeaveGame
```mermaid
sequenceDiagram

    participant C1 as Client1
    participant S as Server
    participant C2 as Client2
    note right of C1: Client1 Clicks Leave
    C1->>S: 1. OpLeave [PlayerID]
    note over S: RoomState: Waiting
    S-->>C1: 3. OpRoomUpdate [all PlayerIDs, RoomState]
    note left of C2: Client2 Enters Waiting
    note right of C1: Client1 in MatchPanel
    note right of C2: Loop until Client2 Clicks Leave
    note left of C2: Client2 Clicks Leave
    C2->>S: 1. OpLeave [PlayerID]
    note over S: RoomState: Empty
    S-->>C2: 3. OpRoomUpdate [all PlayerIDs, RoomState]
    note over S: Room Deleted