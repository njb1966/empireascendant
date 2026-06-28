# Federated Terminal Game Network Design Guide

## Introduction
This guide provides essential information for designing a new game compatible with the **InterDOOR** federated terminal game network. It covers key concepts, design considerations, and technical requirements to ensure seamless integration with the network.

---

### 1. Overview of InterDOOR
**InterDOOR** is an open protocol and reference implementation that allows independently operated nodes to host terminal-based multiplayer games with shared state, cross-node player identity, and synchronized gameplay across the network. Empire Ascendant is the game described by this project folder. Other InterDoor games are examples only and do not define this project.

---

### 2. Key Concepts

#### **Nodes**
- **SSH Game Servers**: Players connect via SSH to interact with the game.
- **Independent Operation**: Each node runs its own SQLite database and serves players independently.
- **Federation**: Nodes can federate (connect to the hub) for cross-node functionality such as wanderers, PvP, travel, and shared debts.

#### **Hub**
- **Centralized Management**: Manages player identities, game states, and facilitates communication between nodes.
- **Public SSH Portal**: Players discover games and connect via an ANSI terminal directory.

---

### 3. Design Considerations

#### **Game Structure**
- **Engine Agnostic**: The engine should define generic interfaces (players, sessions, actions, state, events) without being tied to specific game mechanics or settings.
- **Modular Implementation**: Games register as modules against the engine API, allowing for multiple games on a single node.

#### **State Management**
- **Shared State**: Ensure that critical game states are synchronized across nodes. This includes player progress, inventory, and shared world events.
- **Conflict Resolution**: Implement mechanisms to handle conflicts arising from simultaneous changes to shared states.

---

### 4. Technical Requirements

#### **Node Configuration**
- **SSH Server Setup**: Configure the SSH server on each node for secure connections.
- **Database Management**: Use SQLite for local data storage, ensuring consistency and efficiency.
- **Federation API**: Implement communication protocols with the hub to sync game state and player identities.

#### **Hub Integration**
- **API Key Management**: Nodes must register with the hub to obtain an API key for federated operations.
- **TLS Termination**: Use a Caddy reverse proxy for TLS termination on nodes.
- **Public Portal Setup**: Ensure nodes are listed in the public SSH portal for discoverability.

---

### 5. Implementation Guidelines

#### **Game Initialization**
- **Node Registration**: Automatically register with the hub upon first run, obtaining an API key.
- **Federation Activation**: Enable federation to start syncing game state and player identities.

#### **Player Management**
- **Account Creation**: Allow players to create accounts via SSH.
- **Identity Synchronization**: Ensure player identities are consistent across nodes during federated play.

#### **Gameplay Mechanics**
- **Shared World Events**: Design world events that occur globally and affect all nodes in the network.
- **Cross-Node Features**: Implement features like wanderers (players moving between nodes) and PvP (player vs. player combat).

---

### 6. Example Node Setup

#### **Node Configuration File (`node.json`)**:
```json
{
  "host": "mynode.example.com",
  "port": 2324,
  "hub_url": "https://hub.interdoor.net",
  "max_sessions": 32,
  "idle_timeout": 600,
  "game_title": "Empire Ascendant"
}
```

#### **First Run Commands**:
```bash
sudo cp dist/interdoor-dominion-linux-amd64 /usr/local/bin/interdoor-dominion
interdoor-dominion -config /etc/interdoor-dominion/node.json
```

---

### 7. Conclusion

By following this guide, you can design and implement a new game that seamlessly integrates with the **InterDOOR** federated terminal game network. Ensure compliance with technical requirements, maintain consistency across nodes, and leverage the hub's capabilities to enhance the player experience.

For further assistance, refer to the official documentation or reach out to the InterDOOR community for support.

---

Feel free to use this guide as a reference when designing your new game to ensure it meets the necessary criteria for compatibility with the **InterDOOR** network.
